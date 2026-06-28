package ports

import "time"

// PrescriptionProjection es la proyección de lectura de una receta (ADR-0022).
type PrescriptionProjection struct {
	EvidenceID          string
	TenantID            string
	PatientID           string
	MedicamentoGenerico string
	Dosis               string
	Diagnostico         string
	Indicaciones        string
	Seguimiento         string
	State               string
	CreatedAt           time.Time
	IssuedAt            *time.Time
}

// PrescriptionProjectionRepository define el contrato de acceso a la
// proyección de recetas (ADR-0022).
type PrescriptionProjectionRepository interface {
	Upsert(p PrescriptionProjection) error
	UpdateState(tenantID, evidenceID, state string) error
	ListByPatient(tenantID, patientID string) ([]PrescriptionProjection, error)
	ListAll(tenantID string) ([]PrescriptionProjection, error)
}
