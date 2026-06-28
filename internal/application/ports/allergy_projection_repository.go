package ports

import "time"

// AllergyProjection es la proyección de lectura de una alergia (ADR-0022).
// Se escribe al crear/void. Se lee para listar. Nunca reemplaza a evidence.
type AllergyProjection struct {
	EvidenceID   string
	TenantID     string
	PatientID    string
	Agente       string
	TipoReaccion string
	Criticidad   string
	Certeza      string
	FechaInicio  string
	Notas        string
	State        string
	CreatedAt    time.Time
	IssuedAt     *time.Time
}

// AllergyProjectionRepository define el contrato de acceso a la proyección
// de alergias (ADR-0022). Solo el AllergyService escribe aquí.
type AllergyProjectionRepository interface {
	// Upsert crea o actualiza la proyección de una alergia.
	Upsert(p AllergyProjection) error
	// UpdateState actualiza solo el estado (para void).
	UpdateState(tenantID, evidenceID, state string) error
	// ListByPatient retorna alergias activas de un paciente.
	ListByPatient(tenantID, patientID string) ([]AllergyProjection, error)
}
