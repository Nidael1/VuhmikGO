package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Patient representa un paciente del consultorio en PostgreSQL.
// Pertenece a los Asteroides (CRM) — no al Core.
// Campos minimos segun NOM-004-SSA3-2012, numeral 5.9.
type Patient struct {
	ID              string
	TenantID        string
	Nombre          string
	FechaNacimiento string
	Sexo            string
	NumExpediente   string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// PatientRepository es el adaptador PostgreSQL para pacientes.
type PatientRepository struct {
	pool *pgxpool.Pool
}

// NewPatientRepository retorna un repositorio de pacientes PostgreSQL.
func NewPatientRepository(pool *pgxpool.Pool) *PatientRepository {
	return &PatientRepository{pool: pool}
}

// Create inserta un nuevo paciente en la BD.
func (r *PatientRepository) Create(p Patient) error {
	sql := `
		INSERT INTO patients (id, tenant_id, nombre, fecha_nacimiento, sexo, num_expediente, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.pool.Exec(context.Background(), sql,
		p.ID, p.TenantID, p.Nombre, p.FechaNacimiento,
		p.Sexo, p.NumExpediente, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("error al crear paciente: %w", err)
	}
	return nil
}

// FindByID recupera un paciente por ID dentro del tenant.
func (r *PatientRepository) FindByID(tenantID, id string) (Patient, error) {
	sql := `
		SELECT id, tenant_id, nombre, fecha_nacimiento, sexo, num_expediente, created_at, updated_at
		FROM patients WHERE id = $1 AND tenant_id = $2`
	row := r.pool.QueryRow(context.Background(), sql, id, tenantID)
	var p Patient
	if err := row.Scan(
		&p.ID, &p.TenantID, &p.Nombre, &p.FechaNacimiento,
		&p.Sexo, &p.NumExpediente, &p.CreatedAt, &p.UpdatedAt,
	); err != nil {
		return Patient{}, fmt.Errorf("paciente no encontrado: %w", err)
	}
	return p, nil
}

// FindAll retorna todos los pacientes del tenant.
func (r *PatientRepository) FindAll(tenantID string) ([]Patient, error) {
	sql := `
		SELECT id, tenant_id, nombre, fecha_nacimiento, sexo, num_expediente, created_at, updated_at
		FROM patients WHERE tenant_id = $1 ORDER BY nombre ASC`
	rows, err := r.pool.Query(context.Background(), sql, tenantID)
	if err != nil {
		return nil, fmt.Errorf("error al listar pacientes: %w", err)
	}
	defer rows.Close()
	var result []Patient
	for rows.Next() {
		var p Patient
		if err := rows.Scan(
			&p.ID, &p.TenantID, &p.Nombre, &p.FechaNacimiento,
			&p.Sexo, &p.NumExpediente, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("error al escanear paciente: %w", err)
		}
		result = append(result, p)
	}
	return result, nil
}

// Update actualiza los datos de un paciente.
func (r *PatientRepository) Update(tenantID string, p Patient) error {
	sql := `
		UPDATE patients
		SET nombre = $1, fecha_nacimiento = $2, sexo = $3, updated_at = $4
		WHERE id = $5 AND tenant_id = $6`
	_, err := r.pool.Exec(context.Background(), sql,
		p.Nombre, p.FechaNacimiento, p.Sexo, p.UpdatedAt, p.ID, tenantID,
	)
	if err != nil {
		return fmt.Errorf("error al actualizar paciente: %w", err)
	}
	return nil
}

// NextExpediente genera el siguiente numero de expediente para el tenant.
func (r *PatientRepository) NextExpediente(tenantID string) (string, error) {
	sql := `SELECT COUNT(1) FROM patients WHERE tenant_id = $1`
	var count int
	if err := r.pool.QueryRow(context.Background(), sql, tenantID).Scan(&count); err != nil {
		return "", fmt.Errorf("error al contar pacientes: %w", err)
	}
	return fmt.Sprintf("EXP-%04d", count+1), nil
}
