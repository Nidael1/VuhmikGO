package ports

import "github.com/Nidael1/VuhmikGO/internal/core/evidence"

// EvidenceRepository es el puerto de acceso a datos de evidencia.
type EvidenceRepository interface {
	Create(e evidence.Evidence) error
	FindByID(tenantID, id string) (evidence.Evidence, error)
	Update(tenantID string, e evidence.Evidence) error
	// UpdateForVoid permite void sin GuardMutation (ADR-0006).
	UpdateForVoid(tenantID string, e evidence.Evidence) error
	FindAll(tenantID string) ([]evidence.Evidence, error)
}
