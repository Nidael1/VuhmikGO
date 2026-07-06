package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Patient representa un paciente del consultorio en PostgreSQL.
type Patient struct {
	CURP            *string
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

func NewPatientRepository(pool *pgxpool.Pool) *PatientRepository {
	return &PatientRepository{pool: pool}
}

func (r *PatientRepository) Create(p Patient) error {
	sql := `
		INSERT INTO patients (id, tenant_id, nombre, fecha_nacimiento, sexo, num_expediente, curp, created_at, updated_at)
		VALUES ($1, $2, $3, $4::date, $5, $6, $7, $8, $9)`
	_, err := r.pool.Exec(context.Background(), sql,
		p.ID, p.TenantID, p.Nombre, p.FechaNacimiento,
		p.Sexo, p.NumExpediente, p.CURP, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("error al crear paciente: %w", err)
	}
	return nil
}

func scanPatient(row interface{ Scan(...any) error }) (Patient, error) {
	var p Patient
	var fechaNac time.Time
	if err := row.Scan(
		&p.ID, &p.TenantID, &p.Nombre, &fechaNac,
		&p.Sexo, &p.NumExpediente, &p.CURP, &p.CreatedAt, &p.UpdatedAt,
	); err != nil {
		return Patient{}, err
	}
	p.FechaNacimiento = fechaNac.Format("2006-01-02")
	return p, nil
}

func (r *PatientRepository) FindByID(tenantID, id string) (Patient, error) {
	sql := `
		SELECT id, tenant_id, nombre, fecha_nacimiento, sexo, num_expediente, curp, created_at, updated_at
		FROM patients WHERE id = $1 AND tenant_id = $2`
	row := r.pool.QueryRow(context.Background(), sql, id, tenantID)
	p, err := scanPatient(row)
	if err != nil {
		return Patient{}, fmt.Errorf("paciente no encontrado: %w", err)
	}
	return p, nil
}

func (r *PatientRepository) FindAll(tenantID string) ([]Patient, error) {
	sql := `
		SELECT id, tenant_id, nombre, fecha_nacimiento, sexo, num_expediente, curp, created_at, updated_at
		FROM patients WHERE tenant_id = $1 ORDER BY nombre ASC`
	rows, err := r.pool.Query(context.Background(), sql, tenantID)
	if err != nil {
		return nil, fmt.Errorf("error al listar pacientes: %w", err)
	}
	defer rows.Close()
	var result []Patient
	for rows.Next() {
		p, err := scanPatient(rows)
		if err != nil {
			return nil, fmt.Errorf("error al escanear paciente: %w", err)
		}
		result = append(result, p)
	}
	return result, nil
}

func (r *PatientRepository) Update(tenantID string, p Patient) error {
	sql := `
		UPDATE patients
		SET nombre = $1, fecha_nacimiento = $2::date, sexo = $3, updated_at = $4
		WHERE id = $5 AND tenant_id = $6`
	_, err := r.pool.Exec(context.Background(), sql,
		p.Nombre, p.FechaNacimiento, p.Sexo, p.UpdatedAt, p.ID, tenantID,
	)
	if err != nil {
		return fmt.Errorf("error al actualizar paciente: %w", err)
	}
	return nil
}

func (r *PatientRepository) NextExpediente(tenantID string) (string, error) {
	var count int
	sql := `SELECT COUNT(1) FROM patients WHERE tenant_id = $1`
	if err := r.pool.QueryRow(context.Background(), sql, tenantID).Scan(&count); err != nil {
		return "", fmt.Errorf("error al contar pacientes: %w", err)
	}
	return fmt.Sprintf("EXP-%04d", count+1), nil
}

// FindByCURP busca un paciente por CURP dentro de un tenant.
// Retorna error si no existe. Usado en traspaso de paciente (ADR-0009).
func (r *PatientRepository) FindByCURP(tenantID, curp string) (Patient, error) {
	sql := `
		SELECT id, tenant_id, nombre, fecha_nacimiento, sexo, num_expediente, curp, created_at, updated_at
		FROM patients WHERE tenant_id = $1 AND curp = $2 LIMIT 1`
	row := r.pool.QueryRow(context.Background(), sql, tenantID, curp)
	p, err := scanPatient(row)
	if err != nil {
		return Patient{}, fmt.Errorf("paciente no encontrado por CURP: %w", err)
	}
	return p, nil
}
