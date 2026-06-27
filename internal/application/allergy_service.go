package application

import (
	"encoding/json"
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
//  5. Core ejecuta o rechaza.
//  6. Asteroide muestra resultado.
type AllergyService struct {
	repo  ports.EvidenceRepository
	caps  ports.CapabilityRepository
	rubro string
}

// NewAllergyService retorna un AllergyService con sus dependencias.
func NewAllergyService(repo ports.EvidenceRepository, caps ports.CapabilityRepository) *AllergyService {
	return &AllergyService{repo: repo, caps: caps, rubro: "medico"}
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
	return e, nil
}

// ListByPatient retorna todas las alergias activas de un paciente.
func (s *AllergyService) ListByPatient(tenantID, patientID string) ([]evidence.Evidence, error) {
	all, err := s.repo.FindAll(tenantID)
	if err != nil {
		return nil, err
	}
	var result []evidence.Evidence
	for _, e := range all {
		if e.SubjectRef != patientID || e.State != evidence.StateIssued {
			continue
		}
		// Filtrar solo registros de tipo "allergy" (ADR-0016 — Core agnostico)
		var blob map[string]any
		if err := json.Unmarshal([]byte(e.Content), &blob); err != nil {
			continue
		}
		if t, ok := blob["type"].(string); !ok || t != "allergy" {
			continue
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
	return voided, nil
}
