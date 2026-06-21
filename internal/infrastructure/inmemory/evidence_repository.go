// Package inmemory provee adaptadores de repositorio en memoria.
// Solo para desarrollo y testing — no para producción.
package inmemory

import (
	"fmt"
	"sync"

	"github.com/Nidael1/VuhmikGO/internal/core/evidence"
)

// EvidenceRepository es una implementación en memoria del puerto EvidenceRepository.
// Thread-safe. No persiste entre reinicios.
type EvidenceRepository struct {
	mu      sync.RWMutex
	records map[string]evidence.Evidence
}

// NewEvidenceRepository retorna un repositorio en memoria vacío.
func NewEvidenceRepository() *EvidenceRepository {
	return &EvidenceRepository{
		records: make(map[string]evidence.Evidence),
	}
}

// Create persiste un registro nuevo en memoria.
func (r *EvidenceRepository) Create(e evidence.Evidence) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.records[e.ID]; exists {
		return fmt.Errorf("registro con ID %s ya existe", e.ID)
	}
	r.records[e.ID] = e
	return nil
}

// FindByID recupera un registro por ID, exigiendo que pertenezca a tenantID.
// Aislamiento multi-tenant (Issue #56): un registro de otro tenant
// retorna el mismo error que "no encontrado".
func (r *EvidenceRepository) FindByID(tenantID, id string) (evidence.Evidence, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.records[id]
	if !ok || e.TenantID != tenantID {
		return evidence.Evidence{}, fmt.Errorf("registro %s no encontrado", id)
	}
	return e, nil
}

// Update actualiza un registro existente, exigiendo que pertenezca a tenantID.
// Rechaza si el estado actual es issued o locked (ER-CORE-001).
func (r *EvidenceRepository) Update(tenantID string, e evidence.Evidence) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	current, ok := r.records[e.ID]
	if !ok || current.TenantID != tenantID {
		return fmt.Errorf("registro %s no encontrado", e.ID)
	}
	if err := evidence.GuardMutation(current); err != nil {
		return err
	}
	r.records[e.ID] = e
	return nil
}

// FindAll retorna todos los registros del tenant dado.
// Nunca retorna registros de otro tenant.
func (r *EvidenceRepository) FindAll(tenantID string) ([]evidence.Evidence, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []evidence.Evidence
	for _, e := range r.records {
		if e.TenantID == tenantID {
			result = append(result, e)
		}
	}
	return result, nil
}
