// Package evidence contiene las entidades y reglas del dominio Core.
// El Core es agnostico: no conoce UI, pais, reglas clinicas ni legales.
package evidence

import "time"

// Evidence representa un registro de evidencia clinica inmutable.
// Una vez emitido, no puede modificarse — solo anularse y reemplazarse.
type Evidence struct {
	ID           string
	TenantID     string
	SubjectID    string
	Notes        string
	State        State
	CreatedAt    time.Time
	IssuedAt     *time.Time
	VoidedAt     *time.Time
	ReplacedByID *string
}
