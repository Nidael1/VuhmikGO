package postgres

import (
	"context"
	"fmt"

	"github.com/Nidael1/VuhmikGO/internal/application/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LabResultProjectionRepository es el adaptador PostgreSQL para proyecciones
// de resultados de laboratorio (ADR-0022).
type LabResultProjectionRepository struct {
	pool *pgxpool.Pool
}

func NewLabResultProjectionRepository(pool *pgxpool.Pool) *LabResultProjectionRepository {
	return &LabResultProjectionRepository{pool: pool}
}

func (r *LabResultProjectionRepository) Upsert(p ports.LabResultProjection) error {
	sql := `
		INSERT INTO lab_result_projections
			(evidence_id, tenant_id, patient_id, estudio, fecha_estudio,
			 resultado, laboratorio, unidades, valor_referencia, notas,
			 state, created_at, issued_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (evidence_id) DO UPDATE SET
			estudio          = EXCLUDED.estudio,
			fecha_estudio    = EXCLUDED.fecha_estudio,
			resultado        = EXCLUDED.resultado,
			laboratorio      = EXCLUDED.laboratorio,
			unidades         = EXCLUDED.unidades,
			valor_referencia = EXCLUDED.valor_referencia,
			notas            = EXCLUDED.notas,
			state            = EXCLUDED.state,
			issued_at        = EXCLUDED.issued_at`
	_, err := r.pool.Exec(context.Background(), sql,
		p.EvidenceID, p.TenantID, p.PatientID,
		p.Estudio, p.FechaEstudio, p.Resultado, p.Laboratorio,
		p.Unidades, p.ValorReferencia, p.Notas, p.State, p.CreatedAt, p.IssuedAt,
	)
	if err != nil {
		return fmt.Errorf("error al guardar proyección de resultado de laboratorio: %w", err)
	}
	return nil
}

func (r *LabResultProjectionRepository) UpdateState(tenantID, evidenceID, state string) error {
	sql := `
		UPDATE lab_result_projections
		SET state = $1
		WHERE evidence_id = $2 AND tenant_id = $3`
	_, err := r.pool.Exec(context.Background(), sql, state, evidenceID, tenantID)
	if err != nil {
		return fmt.Errorf("error al actualizar estado de proyección: %w", err)
	}
	return nil
}

func (r *LabResultProjectionRepository) ListByPatient(tenantID, patientID string) ([]ports.LabResultProjection, error) {
	sql := `
		SELECT evidence_id, tenant_id, patient_id, estudio, fecha_estudio,
		       resultado, laboratorio, unidades, valor_referencia, notas,
		       state, created_at, issued_at
		FROM lab_result_projections
		WHERE tenant_id = $1 AND patient_id = $2 AND state = 'issued'
		ORDER BY created_at DESC`
	rows, err := r.pool.Query(context.Background(), sql, tenantID, patientID)
	if err != nil {
		return nil, fmt.Errorf("error al listar proyecciones de resultados de laboratorio: %w", err)
	}
	defer rows.Close()
	var result []ports.LabResultProjection
	for rows.Next() {
		var p ports.LabResultProjection
		if err := rows.Scan(
			&p.EvidenceID, &p.TenantID, &p.PatientID,
			&p.Estudio, &p.FechaEstudio, &p.Resultado, &p.Laboratorio,
			&p.Unidades, &p.ValorReferencia, &p.Notas, &p.State, &p.CreatedAt, &p.IssuedAt,
		); err != nil {
			return nil, fmt.Errorf("error al escanear proyección: %w", err)
		}
		result = append(result, p)
	}
	return result, nil
}
