package postgres

import (
	"context"
	"fmt"

	"github.com/Nidael1/VuhmikGO/internal/application/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DiagnosisProjectionRepository es el adaptador PostgreSQL para proyecciones
// de diagnósticos (ADR-0022).
type DiagnosisProjectionRepository struct {
	pool *pgxpool.Pool
}

func NewDiagnosisProjectionRepository(pool *pgxpool.Pool) *DiagnosisProjectionRepository {
	return &DiagnosisProjectionRepository{pool: pool}
}

func (r *DiagnosisProjectionRepository) Upsert(p ports.DiagnosisProjection) error {
	sql := `
		INSERT INTO diagnosis_projections
			(evidence_id, tenant_id, patient_id, descripcion, codigo_cie10,
			 tipo, estado_problema, fecha_inicio, notas, state, created_at, issued_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (evidence_id) DO UPDATE SET
			descripcion     = EXCLUDED.descripcion,
			codigo_cie10    = EXCLUDED.codigo_cie10,
			tipo            = EXCLUDED.tipo,
			estado_problema = EXCLUDED.estado_problema,
			fecha_inicio    = EXCLUDED.fecha_inicio,
			notas           = EXCLUDED.notas,
			state           = EXCLUDED.state,
			issued_at       = EXCLUDED.issued_at`
	_, err := r.pool.Exec(context.Background(), sql,
		p.EvidenceID, p.TenantID, p.PatientID,
		p.Descripcion, p.CodigoCIE10, p.Tipo, p.EstadoProblema,
		p.FechaInicio, p.Notas, p.State, p.CreatedAt, p.IssuedAt,
	)
	if err != nil {
		return fmt.Errorf("error al guardar proyección de diagnóstico: %w", err)
	}
	return nil
}

func (r *DiagnosisProjectionRepository) UpdateState(tenantID, evidenceID, state string) error {
	sql := `
		UPDATE diagnosis_projections
		SET state = $1
		WHERE evidence_id = $2 AND tenant_id = $3`
	_, err := r.pool.Exec(context.Background(), sql, state, evidenceID, tenantID)
	if err != nil {
		return fmt.Errorf("error al actualizar estado de proyección: %w", err)
	}
	return nil
}

func (r *DiagnosisProjectionRepository) ListByPatient(tenantID, patientID string) ([]ports.DiagnosisProjection, error) {
	sql := `
		SELECT evidence_id, tenant_id, patient_id, descripcion, codigo_cie10,
		       tipo, estado_problema, fecha_inicio, notas, state, created_at, issued_at
		FROM diagnosis_projections
		WHERE tenant_id = $1 AND patient_id = $2 AND state = 'issued'
		ORDER BY created_at DESC`
	rows, err := r.pool.Query(context.Background(), sql, tenantID, patientID)
	if err != nil {
		return nil, fmt.Errorf("error al listar proyecciones de diagnósticos: %w", err)
	}
	defer rows.Close()
	var result []ports.DiagnosisProjection
	for rows.Next() {
		var p ports.DiagnosisProjection
		if err := rows.Scan(
			&p.EvidenceID, &p.TenantID, &p.PatientID,
			&p.Descripcion, &p.CodigoCIE10, &p.Tipo, &p.EstadoProblema,
			&p.FechaInicio, &p.Notas, &p.State, &p.CreatedAt, &p.IssuedAt,
		); err != nil {
			return nil, fmt.Errorf("error al escanear proyección: %w", err)
		}
		result = append(result, p)
	}
	return result, nil
}
