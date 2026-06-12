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
//
// Aislamiento multi-tenant (Issue #56): todas las operaciones que
// recuperan o mutan un registro existente exigen tenantID explícito,
// propagado al repositorio. Un tenant nunca puede leer ni mutar
// registros de otro tenant, incluso conociendo el ID exacto.
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
// Exige tenantID — solo opera sobre registros del tenant dado.
func (s *ECEService) Issue(tenantID, id string) (evidence.Evidence, error) {
	e, err := s.repo.FindByID(tenantID, id)
	if err != nil {
		return evidence.Evidence{}, err
	}
	if err := evidence.GuardTransition(e.State, evidence.StateIssued); err != nil {
		return evidence.Evidence{}, err
	}
	now := time.Now().UTC()
	e.State = evidence.StateIssued
	e.IssuedAt = &now
	if err := s.repo.Update(tenantID, e); err != nil {
		return evidence.Evidence{}, err
	}
	return e, nil
}

// Lock bloquea un registro issued, transitando a locked.
// Post-lock el registro es completamente inmutable.
// Exige tenantID — solo opera sobre registros del tenant dado.
func (s *ECEService) Lock(tenantID, id string) (evidence.Evidence, error) {
	e, err := s.repo.FindByID(tenantID, id)
	if err != nil {
		return evidence.Evidence{}, err
	}
	if err := evidence.GuardTransition(e.State, evidence.StateLocked); err != nil {
		return evidence.Evidence{}, err
	}
	e.State = evidence.StateLocked
	if err := s.repo.Update(tenantID, e); err != nil {
		return evidence.Evidence{}, err
	}
	return e, nil
}

// IssueAndLock emite y bloquea en una sola operación atómica de dominio.
// Equivale a draft → issued → locked.
// Post-operación el registro es completamente inmutable.
func (s *ECEService) IssueAndLock(tenantID, id string) (evidence.Evidence, error) {
	issued, err := s.Issue(tenantID, id)
	if err != nil {
		return evidence.Evidence{}, err
	}
	return s.Lock(tenantID, issued.ID)
}

// Void anula un registro con reason_code obligatorio.
// El registro original queda voided y preservado en historial.
// Rechaza si reason_code no está en el catálogo.
// Exige tenantID — solo opera sobre registros del tenant dado.
func (s *ECEService) Void(tenantID, id string, reasonCode evidence.ReasonCode) (evidence.Evidence, error) {
	e, err := s.repo.FindByID(tenantID, id)
	if err != nil {
		return evidence.Evidence{}, err
	}
	voided, err := evidence.Void(e, reasonCode, time.Now().UTC())
	if err != nil {
		return evidence.Evidence{}, err
	}
	if err := s.repo.Update(tenantID, voided); err != nil {
		return evidence.Evidence{}, err
	}
	return voided, nil
}

// Replace anula el original y crea el reemplazo emitido en una operación.
// El original queda voided. El reemplazo queda issued con replaced_by_id enlazado.
// Rechaza si el original no está en estado que permita void.
// El original debe pertenecer a tenantID; el reemplazo se crea en el
// mismo tenant (no se permite mover registros entre tenants).
func (s *ECEService) Replace(
	tenantID string,
	originalID string,
	replacementID string,
	reasonCode evidence.ReasonCode,
) (original evidence.Evidence, replacement evidence.Evidence, err error) {
	orig, err := s.repo.FindByID(tenantID, originalID)
	if err != nil {
		return evidence.Evidence{}, evidence.Evidence{}, err
	}

	repl := evidence.Evidence{
		ID:        replacementID,
		TenantID:  tenantID,
		State:     evidence.StateDraft,
		CreatedAt: time.Now().UTC(),
	}

	voided, issued, err := evidence.Replace(orig, repl, reasonCode, time.Now().UTC())
	if err != nil {
		return evidence.Evidence{}, evidence.Evidence{}, err
	}

	if err := s.repo.Update(tenantID, voided); err != nil {
		return evidence.Evidence{}, evidence.Evidence{}, err
	}
	if err := s.repo.Create(issued); err != nil {
		return evidence.Evidence{}, evidence.Evidence{}, err
	}
	return voided, issued, nil
}
