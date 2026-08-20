package postgres

import (
	"context"
	"fmt"

	"github.com/Nidael1/VuhmikGO/internal/application/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AppointmentRepository struct {
	pool *pgxpool.Pool
}

func NewAppointmentRepository(pool *pgxpool.Pool) *AppointmentRepository {
	return &AppointmentRepository{pool: pool}
}

func (r *AppointmentRepository) Create(a ports.Appointment) error {
	sql := `
		INSERT INTO appointments
			(id, tenant_id, patient_id, scheduled_at, duration_min, reason, state,
			 consultation_id, notes, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`
	consultationID := &a.ConsultationID
	if a.ConsultationID == "" {
		consultationID = nil
	}
	_, err := r.pool.Exec(context.Background(), sql,
		a.ID, a.TenantID, a.PatientID, a.ScheduledAt, a.DurationMin,
		a.Reason, a.State, consultationID, a.Notes, a.CreatedAt, a.UpdatedAt,
	)
	return err
}

func (r *AppointmentRepository) ListByTenant(tenantID string) ([]ports.Appointment, error) {
	sql := `
		SELECT id, tenant_id, patient_id, scheduled_at, duration_min, reason, state,
		       COALESCE(consultation_id,''), notes, created_at, updated_at
		FROM appointments
		WHERE tenant_id = $1
		ORDER BY scheduled_at DESC`
	rows, err := r.pool.Query(context.Background(), sql, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAppointments(rows)
}

func (r *AppointmentRepository) ListToday(tenantID, date string) ([]ports.Appointment, error) {
	sql := `
		SELECT id, tenant_id, patient_id, scheduled_at, duration_min, reason, state,
		       COALESCE(consultation_id,''), notes, created_at, updated_at
		FROM appointments
		WHERE tenant_id = $1
		  AND scheduled_at::date = $2::date
		ORDER BY scheduled_at ASC`
	rows, err := r.pool.Query(context.Background(), sql, tenantID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAppointments(rows)
}

func (r *AppointmentRepository) ListByPatient(tenantID, patientID string) ([]ports.Appointment, error) {
	sql := `
		SELECT id, tenant_id, patient_id, scheduled_at, duration_min, reason, state,
		       COALESCE(consultation_id,''), notes, created_at, updated_at
		FROM appointments
		WHERE tenant_id = $1 AND patient_id = $2
		ORDER BY scheduled_at DESC`
	rows, err := r.pool.Query(context.Background(), sql, tenantID, patientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAppointments(rows)
}

func (r *AppointmentRepository) FindByID(tenantID, id string) (ports.Appointment, error) {
	sql := `
		SELECT id, tenant_id, patient_id, scheduled_at, duration_min, reason, state,
		       COALESCE(consultation_id,''), notes, created_at, updated_at
		FROM appointments
		WHERE tenant_id = $1 AND id = $2`
	var a ports.Appointment
	err := r.pool.QueryRow(context.Background(), sql, tenantID, id).Scan(
		&a.ID, &a.TenantID, &a.PatientID, &a.ScheduledAt, &a.DurationMin,
		&a.Reason, &a.State, &a.ConsultationID, &a.Notes, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return ports.Appointment{}, fmt.Errorf("appointment not found: %w", err)
	}
	return a, nil
}

func (r *AppointmentRepository) UpdateState(tenantID, id, state string) error {
	sql := `UPDATE appointments SET state = $1, updated_at = now() WHERE tenant_id = $2 AND id = $3`
	_, err := r.pool.Exec(context.Background(), sql, state, tenantID, id)
	return err
}

func (r *AppointmentRepository) SetConsultation(tenantID, id, consultationID string) error {
	sql := `UPDATE appointments SET consultation_id = $1, state = 'completed', updated_at = now()
	        WHERE tenant_id = $2 AND id = $3`
	_, err := r.pool.Exec(context.Background(), sql, consultationID, tenantID, id)
	return err
}

func scanAppointments(rows interface{ Next() bool; Scan(...any) error }) ([]ports.Appointment, error) {
	var result []ports.Appointment
	for rows.Next() {
		var a ports.Appointment
		if err := rows.Scan(
			&a.ID, &a.TenantID, &a.PatientID, &a.ScheduledAt, &a.DurationMin,
			&a.Reason, &a.State, &a.ConsultationID, &a.Notes, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, nil
}
