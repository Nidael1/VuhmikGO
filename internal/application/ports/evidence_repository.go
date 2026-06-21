// Package ports define las interfaces de aplicación para acceso a datos.
// Las implementaciones concretas viven en internal/infrastructure.
package ports

import (
	"github.com/Nidael1/VuhmikGO/internal/core/evidence"
)

// EvidenceRepository es el puerto de acceso a datos de evidencia.
// Toda persistencia de evidencia debe pasar por esta interfaz.
//
// Aislamiento multi-tenant: FindByID, Update y FindAll
// exigen tenantID explícito y filtran por él.
type EvidenceRepository interface {
	// Create persiste un registro Evidence en estado draft.
	Create(e evidence.Evidence) error

	// FindByID recupera un registro por ID exigiendo que pertenezca al tenant.
	FindByID(tenantID, id string) (evidence.Evidence, error)

	// Update persiste cambios de estado exigiendo que pertenezca al tenant.
	// Rechaza si el estado actual es issued o locked (ER-CORE-001).
	Update(tenantID string, e evidence.Evidence) error

	// FindAll retorna todos los registros del tenant dado.
	// Nunca retorna registros de otro tenant.
	FindAll(tenantID string) ([]evidence.Evidence, error)
}
