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

// FindByID recupera un registro por ID.
func (r *EvidenceRepository) FindByID(id string) (evidence.Evidence, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.records[id]
	if !ok {
		return evidence.Evidence{}, fmt.Errorf("registro %s no encontrado", id)
	}
	return e, nil
}

// Update actualiza un registro existente.
// Rechaza si el estado actual es issued o locked (ER-CORE-001).
func (r *EvidenceRepository) Update(e evidence.Evidence) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	current, ok := r.records[e.ID]
	if !ok {
		return fmt.Errorf("registro %s no encontrado", e.ID)
	}
	if err := evidence.GuardMutation(current); err != nil {
		return err
	}
	r.records[e.ID] = e
	return nil
}
