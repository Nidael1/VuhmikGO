package ports

import "time"

// Appointment representa una cita agendada (scheduler_ui, ADR-0031).
type Appointment struct {
	ID             string
	TenantID       string
	PatientID      string
	ScheduledAt    time.Time
	DurationMin    int
	Reason         string
	State          string
	ConsultationID string
	Notes          string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// AppointmentRepository define el contrato de acceso a la tabla appointments.
type AppointmentRepository interface {
	Create(a Appointment) error
	ListByTenant(tenantID string) ([]Appointment, error)
	ListToday(tenantID, date string) ([]Appointment, error)
	ListByPatient(tenantID, patientID string) ([]Appointment, error)
	FindByID(tenantID, id string) (Appointment, error)
	UpdateState(tenantID, id, state string) error
	SetConsultation(tenantID, id, consultationID string) error
}
