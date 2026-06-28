package application

import (
	"fmt"
	"time"

	"github.com/Nidael1/VuhmikGO/internal/application/ports"
	"github.com/Nidael1/VuhmikGO/internal/core/evidence"
	"github.com/Nidael1/VuhmikGO/internal/shaders"
)

// AllergyService orquesta el caso de uso de alergias e intolerancias.
//
// Flujo correcto segun la documentacion:
//  1. Handler recibe y delega.
//  2. AllergyService prepara la operacion.
//  3. Core consulta politica al Shader (CapabilityGuard + AllergyShader).
//  4. Shader responde allow/deny — no muta, no ejecuta.
//  5. Core ejecuta o rechaza + proyeccion se actualiza (ADR-0022).
//  6. Asteroide muestra resultado.
type AllergyService struct {
	repo  ports.EvidenceRepository
	proj  ports.AllergyProjectionRepository
	caps  ports.CapabilityRepository
	rubro string
}

// NewAllergyService retorna un AllergyService con sus dependencias.
func NewAllergyService(repo ports.EvidenceRepository, proj ports.AllergyProjectionRepository, caps ports.CapabilityRepository) *AllergyService {
	return &AllergyService{repo: repo, proj: proj, caps: caps, rubro: "medico"}
}

// Create crea un registro de alergia inmutable en el Core.
func (s *AllergyService) Create(tenantID, actorID, patientID string, content shaders.AllergyContent) (evidence.Evidence, error) {
	inner := shaders.NewAllergyShader()
	guard := shaders.NewCapabilityGuard(inner, s.caps, "allergy", s.rubro)

	if err := shaders.ValidateAllergyContent(content); err != nil {
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

	blob, err := shaders.BuildAllergyBlob(content)
	if err != nil {
		return evidence.Evidence{}, fmt.Errorf("error al construir blob: %w", err)
	}

	now := time.Now().UTC()
	id := fmt.Sprintf("alg-%s-%s", tenantID[:4], now.Format("20060102150405.000"))
	e := evidence.Evidence{
		ID:         id,
		TenantID:   tenantID,
		SubjectRef: patientID,
		Content:    blob,
		State:      evidence.StateDraft,
		CreatedAt:  now,
	}
	if err := s.repo.Create(e); err != nil {
		return evidence.Evidence{}, fmt.Errorf("error al crear alergia: %w", err)
	}

	e.State = evidence.StateIssued
	e.IssuedAt = &now
	if err := s.repo.Update(tenantID, e); err != nil {
		return evidence.Evidence{}, fmt.Errorf("error al emitir alergia: %w", err)
	}

	// Escribir proyección de lectura (ADR-0022)
	proj := ports.AllergyProjection{
		EvidenceID:   e.ID,
		TenantID:     tenantID,
		PatientID:    patientID,
		Agente:       content.Agente,
		TipoReaccion: content.TipoReaccion,
		Criticidad:   content.Criticidad,
		Certeza:      content.Certeza,
		FechaInicio:  content.FechaInicio,
		Notas:        content.Notas,
		State:        string(e.State),
		CreatedAt:    now,
		IssuedAt:     e.IssuedAt,
	}
	if err := s.proj.Upsert(proj); err != nil {
		return evidence.Evidence{}, fmt.Errorf("error al proyectar alergia: %w", err)
	}
	return e, nil
}

// ListByPatient retorna todas las alergias activas de un paciente.
// Lee de allergy_projections (ADR-0022) — O(log n), sin parseo de blobs.
func (s *AllergyService) ListByPatient(tenantID, patientID string) ([]evidence.Evidence, error) {
	projs, err := s.proj.ListByPatient(tenantID, patientID)
	if err != nil {
		return nil, err
	}
	// Reconstruir evidence desde proyección para mantener compatibilidad con handlers
	result := make([]evidence.Evidence, 0, len(projs))
	for _, p := range projs {
		blob, _ := shaders.BuildAllergyBlob(shaders.AllergyContent{
			Agente:       p.Agente,
			TipoReaccion: p.TipoReaccion,
			Criticidad:   p.Criticidad,
			Certeza:      p.Certeza,
			FechaInicio:  p.FechaInicio,
			Notas:        p.Notas,
		})
		e := evidence.Evidence{
			ID:         p.EvidenceID,
			TenantID:   p.TenantID,
			SubjectRef: p.PatientID,
			Content:    blob,
			State:      evidence.State(p.State),
			CreatedAt:  p.CreatedAt,
			IssuedAt:   p.IssuedAt,
		}
		result = append(result, e)
	}
	return result, nil
}

// Void anula una alergia. Correccion via void (ADR-0012).
func (s *AllergyService) Void(tenantID, actorID, allergyID string) (evidence.Evidence, error) {
	e, err := s.repo.FindByID(tenantID, allergyID)
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
	// Actualizar estado en proyección (ADR-0022)
	if err := s.proj.UpdateState(tenantID, allergyID, string(evidence.StateVoided)); err != nil {
		return evidence.Evidence{}, fmt.Errorf("error al actualizar proyección: %w", err)
	}
	return voided, nil
}
