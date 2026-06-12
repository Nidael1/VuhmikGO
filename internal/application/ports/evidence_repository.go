// Package ports define las interfaces de aplicación para acceso a datos.
// Las implementaciones concretas viven en internal/infrastructure.
package ports

import (
	"github.com/Nidael1/VuhmikGO/internal/core/evidence"
)

// EvidenceRepository es el puerto de acceso a datos de evidencia.
// Toda persistencia de evidencia debe pasar por esta interfaz.
// No existe acceso directo a la base de datos desde el Core ni de Shaders.
//
// Aislamiento multi-tenant (Issue #56):
// FindByID y Update requieren tenantID explícito y filtran por él.
// Un registro de un tenant nunca es accesible ni mutable desde otro
// tenant, incluso conociendo el ID exacto del registro.
type EvidenceRepository interface {
	// Create persiste un registro Evidence en estado draft.
	// Retorna error si el ID ya existe o si la estructura es inválida.
	Create(e evidence.Evidence) error

	// FindByID recupera un registro Evidence por su ID, exigiendo que
	// pertenezca al tenantID dado. Retorna error si no existe o si
	// pertenece a otro tenant (sin distinguir ambos casos al caller).
	FindByID(tenantID, id string) (evidence.Evidence, error)

	// Update persiste cambios de estado en un registro existente,
	// exigiendo que pertenezca al tenantID dado.
	// Rechaza UPDATE si el estado actual es issued o locked (ER-CORE-001).
	Update(tenantID string, e evidence.Evidence) error
}
