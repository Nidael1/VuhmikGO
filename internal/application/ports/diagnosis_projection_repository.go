package ports

import "time"

// DiagnosisProjection es la proyección de lectura de un diagnóstico (ADR-0022).
// Se escribe al crear/void. Se lee para listar. Nunca reemplaza a evidence.
type DiagnosisProjection struct {
	EvidenceID     string
	TenantID       string
	PatientID      string
	Descripcion    string
	CodigoCIE10    string
	Tipo           string
	EstadoProblema string
	FechaInicio    string
	Notas          string
	State          string
	CreatedAt      time.Time
	IssuedAt       *time.Time
}

// DiagnosisProjectionRepository define el contrato de acceso a la proyección
// de diagnósticos (ADR-0022). Solo el DiagnosisService escribe aquí.
type DiagnosisProjectionRepository interface {
	// Upsert crea o actualiza la proyección de un diagnóstico.
	Upsert(p DiagnosisProjection) error
	// UpdateState actualiza solo el estado (para void).
	UpdateState(tenantID, evidenceID, state string) error
	// ListByPatient retorna diagnósticos activos de un paciente.
	ListByPatient(tenantID, patientID string) ([]DiagnosisProjection, error)
}
