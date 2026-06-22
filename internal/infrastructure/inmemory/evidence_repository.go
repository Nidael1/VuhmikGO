package inmemory

import (
	"fmt"
	"sync"

	"github.com/Nidael1/VuhmikGO/internal/core/evidence"
)

type EvidenceRepository struct {
	mu      sync.RWMutex
	records map[string]evidence.Evidence
}

func NewEvidenceRepository() *EvidenceRepository {
	return &EvidenceRepository{records: make(map[string]evidence.Evidence)}
}

func (r *EvidenceRepository) Create(e evidence.Evidence) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.records[e.ID]; exists {
		return fmt.Errorf("registro con ID %s ya existe", e.ID)
	}
	r.records[e.ID] = e
	return nil
}

func (r *EvidenceRepository) FindByID(tenantID, id string) (evidence.Evidence, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.records[id]
	if !ok || e.TenantID != tenantID {
		return evidence.Evidence{}, fmt.Errorf("registro %s no encontrado", id)
	}
	return e, nil
}

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

// UpdateForVoid permite actualizar el estado a voided sin GuardMutation.
// Solo para void+replace silencioso (ADR-0006).
func (r *EvidenceRepository) UpdateForVoid(tenantID string, e evidence.Evidence) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	current, ok := r.records[e.ID]
	if !ok || current.TenantID != tenantID {
		return fmt.Errorf("registro %s no encontrado", e.ID)
	}
	r.records[e.ID] = e
	return nil
}

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
