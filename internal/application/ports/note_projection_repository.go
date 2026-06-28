package ports

import "time"

// NoteProjection es la proyección de lectura de una nota clínica (ADR-0022).
type NoteProjection struct {
	EvidenceID string
	TenantID   string
	PatientID  string
	Text       string
	State      string
	CreatedAt  time.Time
	IssuedAt   *time.Time
	// Signos vitales (opcionales)
	TA    string
	FC    string
	FR    string
	Temp  string
	Peso  string
	Talla string
	SAO2  string
}

// NoteProjectionRepository define el contrato de acceso a la proyección
// de notas clínicas (ADR-0022). Solo los handlers de evidencia escriben aquí.
type NoteProjectionRepository interface {
	Upsert(p NoteProjection) error
	UpdateState(tenantID, evidenceID, state string) error
	ListByPatient(tenantID, patientID string) ([]NoteProjection, error)
}
