package application

import (
	"fmt"
	"time"

	"github.com/Nidael1/VuhmikGO/internal/application/ports"
	"github.com/Nidael1/VuhmikGO/internal/core/evidence"
	"github.com/Nidael1/VuhmikGO/internal/shaders"
)

// LabResultService orquesta el caso de uso de resultados de laboratorio
// (ADR-0015).
//
// Flujo correcto segun la documentacion:
//  1. Handler recibe y delega.
//  2. LabResultService prepara la operacion.
//  3. Core consulta politica al Shader (CapabilityGuard + LabResultShader).
//  4. Shader responde allow/deny — no muta, no ejecuta.
//  5. Core ejecuta o rechaza + proyeccion se actualiza (ADR-0022).
//  6. Asteroide muestra resultado.
type LabResultService struct {
	repo  ports.EvidenceRepository
	proj  ports.LabResultProjectionRepository
	caps  ports.CapabilityRepository
	rubro string
}

// NewLabResultService retorna un LabResultService con sus dependencias.
func NewLabResultService(repo ports.EvidenceRepository, proj ports.LabResultProjectionRepository, caps ports.CapabilityRepository) *LabResultService {
	return &LabResultService{repo: repo, proj: proj, caps: caps, rubro: "medico"}
}

// Create crea un registro de resultado de laboratorio inmutable en el Core.
func (s *LabResultService) Create(tenantID, actorID, patientID string, content shaders.LabResultContent) (evidence.Evidence, error) {
	inner := shaders.NewLabResultShader()
	guard := shaders.NewCapabilityGuard(inner, s.caps, "lab_result", s.rubro)

	if err := shaders.ValidateLabResultContent(content); err != nil {
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

	blob, err := shaders.BuildLabResultBlob(content)
	if err != nil {
		return evidence.Evidence{}, fmt.Errorf("error al construir blob: %w", err)
	}

	now := time.Now().UTC()
	id := fmt.Sprintf("lab-%s-%s", tenantID[:4], now.Format("20060102150405.000"))
	e := evidence.Evidence{
		ID:         id,
		TenantID:   tenantID,
		SubjectRef: patientID,
		Content:    blob,
		State:      evidence.StateDraft,
		CreatedAt:  now,
	}
	if err := s.repo.Create(e); err != nil {
		return evidence.Evidence{}, fmt.Errorf("error al crear resultado de laboratorio: %w", err)
	}

	e.State = evidence.StateIssued
	e.IssuedAt = &now
	if err := s.repo.Update(tenantID, e); err != nil {
		return evidence.Evidence{}, fmt.Errorf("error al emitir resultado de laboratorio: %w", err)
	}

	// Escribir proyección de lectura (ADR-0022)
	proj := ports.LabResultProjection{
		EvidenceID:      e.ID,
		TenantID:        tenantID,
		PatientID:       patientID,
		Estudio:         content.Estudio,
		FechaEstudio:    content.FechaEstudio,
		Resultado:       content.Resultado,
		Laboratorio:     content.Laboratorio,
		Unidades:        content.Unidades,
		ValorReferencia: content.ValorReferencia,
		Notas:           content.Notas,
		State:           string(e.State),
		CreatedAt:       now,
		IssuedAt:        e.IssuedAt,
	}
	if err := s.proj.Upsert(proj); err != nil {
		return evidence.Evidence{}, fmt.Errorf("error al proyectar resultado de laboratorio: %w", err)
	}
	return e, nil
}

// ListByPatient retorna todos los resultados activos de un paciente.
// Lee de lab_result_projections (ADR-0022) — sin parseo de blobs.
func (s *LabResultService) ListByPatient(tenantID, patientID string) ([]ports.LabResultProjection, error) {
	return s.proj.ListByPatient(tenantID, patientID)
}

// Void anula un resultado de laboratorio. Corrección via void (ADR-0006).
func (s *LabResultService) Void(tenantID, actorID, labResultID string) (evidence.Evidence, error) {
	e, err := s.repo.FindByID(tenantID, labResultID)
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
	if err := s.proj.UpdateState(tenantID, labResultID, string(evidence.StateVoided)); err != nil {
		return evidence.Evidence{}, fmt.Errorf("error al actualizar proyección: %w", err)
	}
	return voided, nil
}
