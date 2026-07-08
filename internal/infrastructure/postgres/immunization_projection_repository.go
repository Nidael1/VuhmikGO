package postgres

import (
	"context"
	"fmt"

	"github.com/Nidael1/VuhmikGO/internal/application/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ImmunizationProjectionRepository es el adaptador PostgreSQL para
// proyecciones de inmunizaciones (ADR-0022).
type ImmunizationProjectionRepository struct {
	pool *pgxpool.Pool
}

func NewImmunizationProjectionRepository(pool *pgxpool.Pool) *ImmunizationProjectionRepository {
	return &ImmunizationProjectionRepository{pool: pool}
}

func (r *ImmunizationProjectionRepository) Upsert(p ports.ImmunizationProjection) error {
	sql := `
		INSERT INTO immunization_projections
			(evidence_id, tenant_id, patient_id, vacuna, fecha_aplicacion,
			 lote, dosis, via, aplicada_por, notas, state, created_at, issued_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (evidence_id) DO UPDATE SET
			vacuna           = EXCLUDED.vacuna,
			fecha_aplicacion = EXCLUDED.fecha_aplicacion,
			lote             = EXCLUDED.lote,
			dosis            = EXCLUDED.dosis,
			via              = EXCLUDED.via,
			aplicada_por     = EXCLUDED.aplicada_por,
			notas            = EXCLUDED.notas,
			state            = EXCLUDED.state,
			issued_at        = EXCLUDED.issued_at`
	_, err := r.pool.Exec(context.Background(), sql,
		p.EvidenceID, p.TenantID, p.PatientID,
		p.Vacuna, p.FechaAplicacion, p.Lote, p.Dosis, p.Via, p.AplicadaPor,
		p.Notas, p.State, p.CreatedAt, p.IssuedAt,
	)
	if err != nil {
		return fmt.Errorf("error al guardar proyección de inmunización: %w", err)
	}
	return nil
}

func (r *ImmunizationProjectionRepository) UpdateState(tenantID, evidenceID, state string) error {
	sql := `
		UPDATE immunization_projections
		SET state = $1
		WHERE evidence_id = $2 AND tenant_id = $3`
	_, err := r.pool.Exec(context.Background(), sql, state, evidenceID, tenantID)
	if err != nil {
		return fmt.Errorf("error al actualizar estado de proyección: %w", err)
	}
	return nil
}

func (r *ImmunizationProjectionRepository) ListByPatient(tenantID, patientID string) ([]ports.ImmunizationProjection, error) {
	sql := `
		SELECT evidence_id, tenant_id, patient_id, vacuna, fecha_aplicacion,
		       lote, dosis, via, aplicada_por, notas, state, created_at, issued_at
		FROM immunization_projections
		WHERE tenant_id = $1 AND patient_id = $2 AND state = 'issued'
		ORDER BY created_at DESC`
	rows, err := r.pool.Query(context.Background(), sql, tenantID, patientID)
	if err != nil {
		return nil, fmt.Errorf("error al listar proyecciones de inmunizaciones: %w", err)
	}
	defer rows.Close()
	var result []ports.ImmunizationProjection
	for rows.Next() {
		var p ports.ImmunizationProjection
		if err := rows.Scan(
			&p.EvidenceID, &p.TenantID, &p.PatientID,
			&p.Vacuna, &p.FechaAplicacion, &p.Lote, &p.Dosis, &p.Via, &p.AplicadaPor,
			&p.Notas, &p.State, &p.CreatedAt, &p.IssuedAt,
		); err != nil {
			return nil, fmt.Errorf("error al escanear proyección: %w", err)
		}
		result = append(result, p)
	}
	return result, nil
}
