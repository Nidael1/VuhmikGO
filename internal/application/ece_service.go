// Package application contiene los casos de uso de la aplicación.
// Orquesta la lógica entre Shaders, Core y repositorio.
// No contiene lógica clínica. No accede a BD directamente.
package application

import (
	"fmt"
	"time"

	"github.com/Nidael1/VuhmikGO/internal/application/ports"
	"github.com/Nidael1/VuhmikGO/internal/core/evidence"
)

// ECEService orquesta los casos de uso del Expediente Clínico Electrónico.
type ECEService struct {
	repo ports.EvidenceRepository
}

// NewECEService retorna un ECEService con el repositorio dado.
func NewECEService(repo ports.EvidenceRepository) *ECEService {
	return &ECEService{repo: repo}
}

// CreateDraft crea un registro de evidencia en estado draft.
// El ID debe ser provisto por el caller.
func (s *ECEService) CreateDraft(id, tenantID string) (evidence.Evidence, error) {
	e := evidence.Evidence{
		ID:        id,
		TenantID:  tenantID,
		State:     evidence.StateDraft,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.repo.Create(e); err != nil {
		return evidence.Evidence{}, fmt.Errorf("error al crear draft: %w", err)
	}
	return e, nil
}

// Issue emite un registro draft, transitando a issued.
// Registra issued_at. El registro queda inmutable post-emisión.
func (s *ECEService) Issue(id string) (evidence.Evidence, error) {
	e, err := s.repo.FindByID(id)
	if err != nil {
		return evidence.Evidence{}, err
	}
	if err := evidence.GuardTransition(e.State, evidence.StateIssued); err != nil {
		return evidence.Evidence{}, err
	}
	now := time.Now().UTC()
	e.State = evidence.StateIssued
	e.IssuedAt = &now
	if err := s.repo.Update(e); err != nil {
		return evidence.Evidence{}, err
	}
	return e, nil
}

// Lock bloquea un registro issued, transitando a locked.
// Post-lock el registro es completamente inmutable.
func (s *ECEService) Lock(id string) (evidence.Evidence, error) {
	e, err := s.repo.FindByID(id)
	if err != nil {
		return evidence.Evidence{}, err
	}
	if err := evidence.GuardTransition(e.State, evidence.StateLocked); err != nil {
		return evidence.Evidence{}, err
	}
	e.State = evidence.StateLocked
	if err := s.repo.Update(e); err != nil {
		return evidence.Evidence{}, err
	}
	return e, nil
}

// IssueAndLock emite y bloquea en una sola operación atómica de dominio.
// Equivale a draft → issued → locked.
// Post-operación el registro es completamente inmutable.
func (s *ECEService) IssueAndLock(id string) (evidence.Evidence, error) {
	issued, err := s.Issue(id)
	if err != nil {
		return evidence.Evidence{}, err
	}
	return s.Lock(issued.ID)
}

// Void anula un registro con reason_code obligatorio.
// El registro original queda voided y preservado en historial.
// Rechaza si reason_code no está en el catálogo.
func (s *ECEService) Void(id string, reasonCode evidence.ReasonCode) (evidence.Evidence, error) {
	e, err := s.repo.FindByID(id)
	if err != nil {
		return evidence.Evidence{}, err
	}
	voided, err := evidence.Void(e, reasonCode, time.Now().UTC())
	if err != nil {
		return evidence.Evidence{}, err
	}
	if err := s.repo.Update(voided); err != nil {
		return evidence.Evidence{}, err
	}
	return voided, nil
}
