package application

import (
	"fmt"
	"time"

	"github.com/Nidael1/VuhmikGO/internal/application/ports"
	"github.com/Nidael1/VuhmikGO/internal/core/evidence"
	"github.com/Nidael1/VuhmikGO/internal/shaders"
)

// ImmunizationService orquesta el caso de uso de inmunizaciones y
// vacunación (ADR-0014).
//
// Flujo correcto segun la documentacion:
//  1. Handler recibe y delega.
//  2. ImmunizationService prepara la operacion.
//  3. Core consulta politica al Shader (CapabilityGuard + ImmunizationShader).
//  4. Shader responde allow/deny — no muta, no ejecuta.
//  5. Core ejecuta o rechaza + proyeccion se actualiza (ADR-0022).
//  6. Asteroide muestra resultado.
type ImmunizationService struct {
	repo  ports.EvidenceRepository
	proj  ports.ImmunizationProjectionRepository
	caps  ports.CapabilityRepository
	rubro string
}

// NewImmunizationService retorna un ImmunizationService con sus dependencias.
func NewImmunizationService(repo ports.EvidenceRepository, proj ports.ImmunizationProjectionRepository, caps ports.CapabilityRepository) *ImmunizationService {
	return &ImmunizationService{repo: repo, proj: proj, caps: caps, rubro: "medico"}
}

// Create crea un registro de inmunización inmutable en el Core.
func (s *ImmunizationService) Create(tenantID, actorID, patientID string, content shaders.ImmunizationContent) (evidence.Evidence, error) {
	inner := shaders.NewImmunizationShader()
	guard := shaders.NewCapabilityGuard(inner, s.caps, "immunization", s.rubro)

	if err := shaders.ValidateImmunizationContent(content); err != nil {
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

	blob, err := shaders.BuildImmunizationBlob(content)
	if err != nil {
		return evidence.Evidence{}, fmt.Errorf("error al construir blob: %w", err)
	}

	now := time.Now().UTC()
	id := fmt.Sprintf("imm-%s-%s", tenantID[:4], now.Format("20060102150405.000"))
	e := evidence.Evidence{
		ID:         id,
		TenantID:   tenantID,
		SubjectRef: patientID,
		Content:    blob,
		State:      evidence.StateDraft,
		CreatedAt:  now,
	}
	if err := s.repo.Create(e); err != nil {
		return evidence.Evidence{}, fmt.Errorf("error al crear inmunizacion: %w", err)
	}

	e.State = evidence.StateIssued
	e.IssuedAt = &now
	if err := s.repo.Update(tenantID, e); err != nil {
		return evidence.Evidence{}, fmt.Errorf("error al emitir inmunizacion: %w", err)
	}

	// Escribir proyección de lectura (ADR-0022)
	proj := ports.ImmunizationProjection{
		EvidenceID:      e.ID,
		TenantID:        tenantID,
		PatientID:       patientID,
		Vacuna:          content.Vacuna,
		FechaAplicacion: content.FechaAplicacion,
		Lote:            content.Lote,
		Dosis:           content.Dosis,
		Via:             content.Via,
		AplicadaPor:     content.AplicadaPor,
		Notas:           content.Notas,
		State:           string(e.State),
		CreatedAt:       now,
		IssuedAt:        e.IssuedAt,
	}
	if err := s.proj.Upsert(proj); err != nil {
		return evidence.Evidence{}, fmt.Errorf("error al proyectar inmunizacion: %w", err)
	}
	return e, nil
}

// ListByPatient retorna todas las inmunizaciones activas de un paciente.
// Lee de immunization_projections (ADR-0022) — sin parseo de blobs.
func (s *ImmunizationService) ListByPatient(tenantID, patientID string) ([]ports.ImmunizationProjection, error) {
	return s.proj.ListByPatient(tenantID, patientID)
}

// Void anula una inmunización. Corrección via void (ADR-0006).
func (s *ImmunizationService) Void(tenantID, actorID, immunizationID string) (evidence.Evidence, error) {
	e, err := s.repo.FindByID(tenantID, immunizationID)
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
	if err := s.proj.UpdateState(tenantID, immunizationID, string(evidence.StateVoided)); err != nil {
		return evidence.Evidence{}, fmt.Errorf("error al actualizar proyección: %w", err)
	}
	return voided, nil
}
