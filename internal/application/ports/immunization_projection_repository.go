package ports

import "time"

// ImmunizationProjection es la proyección de lectura de una vacuna (ADR-0022).
// Se escribe al crear/void. Se lee para listar. Nunca reemplaza a evidence.
type ImmunizationProjection struct {
	EvidenceID      string
	TenantID        string
	PatientID       string
	Vacuna          string
	FechaAplicacion string
	Lote            string
	Dosis           string
	Via             string
	AplicadaPor     string
	Notas           string
	State           string
	CreatedAt       time.Time
	IssuedAt        *time.Time
}

// ImmunizationProjectionRepository define el contrato de acceso a la
// proyección de inmunizaciones (ADR-0022). Solo el ImmunizationService
// escribe aquí.
type ImmunizationProjectionRepository interface {
	// Upsert crea o actualiza la proyección de una vacuna.
	Upsert(p ImmunizationProjection) error
	// UpdateState actualiza solo el estado (para void).
	UpdateState(tenantID, evidenceID, state string) error
	// ListByPatient retorna vacunas activas de un paciente.
	ListByPatient(tenantID, patientID string) ([]ImmunizationProjection, error)
}
