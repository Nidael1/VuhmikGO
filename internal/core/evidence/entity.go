// Package evidence contiene las entidades y reglas del dominio Core.
// El Core es agnostico: no conoce UI, pais, reglas clinicas ni legales.
// El campo Content es un blob opaco (JSON). El Core nunca lo interpreta.
// El discriminador de tipo vive DENTRO del blob; solo el Shader lo lee.
package evidence

import "time"

// Evidence representa un registro inmutable del Core.
// Una vez emitido, no puede modificarse — solo anularse y reemplazarse.
type Evidence struct {
	ID           string
	TenantID     string
	SubjectRef   string // clave de correlacion opaca (no FK a patients)
	Content      string // blob JSON opaco; el Core nunca lo parsea
	State        State
	CreatedAt    time.Time
	IssuedAt     *time.Time
	VoidedAt     *time.Time
	ReplacedByID *string
}
