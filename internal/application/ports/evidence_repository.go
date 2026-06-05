// Package ports define las interfaces de aplicación para acceso a datos.
// Las implementaciones concretas viven en internal/infrastructure.
package ports

import (
	"github.com/Nidael1/VuhmikGO/internal/core/evidence"
)

// EvidenceRepository es el puerto de acceso a datos de evidencia.
// Toda persistencia de evidencia debe pasar por esta interfaz.
// No existe acceso directo a la base de datos desde el Core ni de Shaders.
type EvidenceRepository interface {
	// Create persiste un registro Evidence en estado draft.
	// Retorna error si el ID ya existe o si la estructura es inválida.
	Create(e evidence.Evidence) error

	// FindByID recupera un registro Evidence por su ID.
	// Retorna error si el registro no existe.
	FindByID(id string) (evidence.Evidence, error)

	// Update persiste cambios de estado en un registro existente.
	// Rechaza UPDATE si el estado actual es issued o locked (ER-CORE-001).
	Update(e evidence.Evidence) error
}
