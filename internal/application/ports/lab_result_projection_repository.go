package ports

import "time"

// LabResultProjection es la proyección de lectura de un resultado de
// laboratorio (ADR-0022). Se escribe al crear/void. Se lee para listar.
// Nunca reemplaza a evidence.
type LabResultProjection struct {
	EvidenceID      string
	TenantID        string
	PatientID       string
	Estudio         string
	FechaEstudio    string
	Resultado       string
	Laboratorio     string
	Unidades        string
	ValorReferencia string
	Notas           string
	State           string
	CreatedAt       time.Time
	IssuedAt        *time.Time
}

// LabResultProjectionRepository define el contrato de acceso a la
// proyección de resultados de laboratorio (ADR-0022). Solo el
// LabResultService escribe aquí.
type LabResultProjectionRepository interface {
	// Upsert crea o actualiza la proyección de un resultado de laboratorio.
	Upsert(p LabResultProjection) error
	// UpdateState actualiza solo el estado (para void).
	UpdateState(tenantID, evidenceID, state string) error
	// ListByPatient retorna resultados activos de un paciente.
	ListByPatient(tenantID, patientID string) ([]LabResultProjection, error)
}
