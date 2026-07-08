package application

import (
	"fmt"
	"time"

	"github.com/Nidael1/VuhmikGO/internal/application/ports"
	"github.com/Nidael1/VuhmikGO/internal/core/evidence"
	"github.com/Nidael1/VuhmikGO/internal/shaders"
)

// DiagnosisService orquesta el caso de uso de diagnósticos estructurados
// y lista de problemas (ADR-0013).
//
// Flujo correcto segun la documentacion:
//  1. Handler recibe y delega.
//  2. DiagnosisService prepara la operacion.
//  3. Core consulta politica al Shader (CapabilityGuard + DiagnosisShader).
//  4. Shader responde allow/deny — no muta, no ejecuta.
//  5. Core ejecuta o rechaza + proyeccion se actualiza (ADR-0022).
//  6. Asteroide muestra resultado.
type DiagnosisService struct {
	repo  ports.EvidenceRepository
	proj  ports.DiagnosisProjectionRepository
	caps  ports.CapabilityRepository
	rubro string
}

// NewDiagnosisService retorna un DiagnosisService con sus dependencias.
func NewDiagnosisService(repo ports.EvidenceRepository, proj ports.DiagnosisProjectionRepository, caps ports.CapabilityRepository) *DiagnosisService {
	return &DiagnosisService{repo: repo, proj: proj, caps: caps, rubro: "medico"}
}

// Create crea un registro de diagnóstico inmutable en el Core.
func (s *DiagnosisService) Create(tenantID, actorID, patientID string, content shaders.DiagnosisContent) (evidence.Evidence, error) {
	inner := shaders.NewDiagnosisShader()
	guard := shaders.NewCapabilityGuard(inner, s.caps, "diagnosis", s.rubro)

	if err := shaders.ValidateDiagnosisContent(content); err != nil {
		return evidence.Evidence{}, fmt.Errorf("contenido invalido: %w", err)
	}

	ctx := shaders.ShaderContext{
		TenantID:  tenantID,
		Operation: shaders.OperationCreate,
		ActorID:   actorID,
	}
	decision := guard.Evaluate(ctx)
	if decision.Result != shaders.DecisionAllow {
		return evidence.Evidence{}, fmt.Errorf("[%s] %s", decision.ErrorCode, decision.Reason)
	}

	blob, err := shaders.BuildDiagnosisBlob(content)
	if err != nil {
		return evidence.Evidence{}, fmt.Errorf("error al construir blob: %w", err)
	}

	now := time.Now().UTC()
	id := fmt.Sprintf("diag-%s-%s", tenantID[:4], now.Format("20060102150405.000"))
	e := evidence.Evidence{
		ID:         id,
		TenantID:   tenantID,
		SubjectRef: patientID,
		Content:    blob,
		State:      evidence.StateDraft,
		CreatedAt:  now,
	}
	if err := s.repo.Create(e); err != nil {
		return evidence.Evidence{}, fmt.Errorf("error al crear diagnostico: %w", err)
	}

	e.State = evidence.StateIssued
	e.IssuedAt = &now
	if err := s.repo.Update(tenantID, e); err != nil {
		return evidence.Evidence{}, fmt.Errorf("error al emitir diagnostico: %w", err)
	}

	// Escribir proyección de lectura (ADR-0022)
	proj := ports.DiagnosisProjection{
		EvidenceID:     e.ID,
		TenantID:       tenantID,
		PatientID:      patientID,
		Descripcion:    content.Descripcion,
		CodigoCIE10:    content.CodigoCIE10,
		Tipo:           content.Tipo,
		EstadoProblema: content.EstadoProblema,
		FechaInicio:    content.FechaInicio,
		Notas:          content.Notas,
		State:          string(e.State),
		CreatedAt:      now,
		IssuedAt:       e.IssuedAt,
	}
	if err := s.proj.Upsert(proj); err != nil {
		return evidence.Evidence{}, fmt.Errorf("error al proyectar diagnostico: %w", err)
	}
	return e, nil
}

// ListByPatient retorna todos los diagnósticos activos de un paciente.
// Lee de diagnosis_projections (ADR-0022) — sin parseo de blobs.
func (s *DiagnosisService) ListByPatient(tenantID, patientID string) ([]ports.DiagnosisProjection, error) {
	return s.proj.ListByPatient(tenantID, patientID)
}

// Void anula un diagnóstico. Corrección via void (ADR-0006).
func (s *DiagnosisService) Void(tenantID, actorID, diagnosisID string) (evidence.Evidence, error) {
	e, err := s.repo.FindByID(tenantID, diagnosisID)
	if err != nil {
		return evidence.Evidence{}, err
	}
	voided, err := evidence.Void(e, evidence.RCVoidErrorDetected, time.Now().UTC())
	if err != nil {
		return evidence.Evidence{}, err
	}
	if err := s.repo.UpdateForVoid(tenantID, voided); err != nil {
		return evidence.Evidence{}, err
	}
	if err := s.proj.UpdateState(tenantID, diagnosisID, string(evidence.StateVoided)); err != nil {
		return evidence.Evidence{}, fmt.Errorf("error al actualizar proyección: %w", err)
	}
	return voided, nil
}
