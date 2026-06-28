package ports

import "time"

// ConsultationProjection es la proyección de lectura de una consulta (ADR-0024).
type ConsultationProjection struct {
	EvidenceID string
	TenantID   string
	PatientID  string
	TA         string
	FC         string
	FR         string
	Temp       string
	Peso       string
	Talla      string
	SAO2       string
	State      string
	CreatedAt  time.Time
	IssuedAt   *time.Time
}

// ConsultationProjectionRepository define el contrato de acceso a proyecciones
// de consultas (ADR-0024).
type ConsultationProjectionRepository interface {
	Upsert(p ConsultationProjection) error
	UpdateState(tenantID, evidenceID, state string) error
	ListByPatient(tenantID, patientID string) ([]ConsultationProjection, error)
	ListAll(tenantID string) ([]ConsultationProjection, error)
	FindByID(tenantID, evidenceID string) (ConsultationProjection, error)
}
