package postgres

import (
	"context"
	"fmt"

	"github.com/Nidael1/VuhmikGO/internal/application/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PrescriptionProjectionRepository es el adaptador PostgreSQL para
// proyecciones de recetas electrónicas (ADR-0022).
type PrescriptionProjectionRepository struct {
	pool *pgxpool.Pool
}

func NewPrescriptionProjectionRepository(pool *pgxpool.Pool) *PrescriptionProjectionRepository {
	return &PrescriptionProjectionRepository{pool: pool}
}

func (r *PrescriptionProjectionRepository) Upsert(p ports.PrescriptionProjection) error {
	sql := `
		INSERT INTO prescription_projections
			(evidence_id, tenant_id, patient_id, medicamento_generico, dosis,
			 diagnostico, indicaciones, seguimiento, state, created_at, issued_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (evidence_id) DO UPDATE SET
			medicamento_generico = EXCLUDED.medicamento_generico,
			dosis                = EXCLUDED.dosis,
			diagnostico          = EXCLUDED.diagnostico,
			indicaciones         = EXCLUDED.indicaciones,
			seguimiento          = EXCLUDED.seguimiento,
			state                = EXCLUDED.state,
			issued_at            = EXCLUDED.issued_at`
	_, err := r.pool.Exec(context.Background(), sql,
		p.EvidenceID, p.TenantID, p.PatientID,
		p.MedicamentoGenerico, p.Dosis,
		p.Diagnostico, p.Indicaciones, p.Seguimiento,
		p.State, p.CreatedAt, p.IssuedAt,
	)
	if err != nil {
		return fmt.Errorf("error al guardar proyección de receta: %w", err)
	}
	return nil
}

func (r *PrescriptionProjectionRepository) UpdateState(tenantID, evidenceID, state string) error {
	sql := `
		UPDATE prescription_projections
		SET state = $1
		WHERE evidence_id = $2 AND tenant_id = $3`
	_, err := r.pool.Exec(context.Background(), sql, state, evidenceID, tenantID)
	return err
}

func (r *PrescriptionProjectionRepository) ListByPatient(tenantID, patientID string) ([]ports.PrescriptionProjection, error) {
	sql := `
		SELECT evidence_id, tenant_id, patient_id, medicamento_generico, dosis,
		       diagnostico, indicaciones, seguimiento, state, created_at, issued_at
		FROM prescription_projections
		WHERE tenant_id = $1 AND patient_id = $2 AND state = 'issued'
		ORDER BY created_at DESC`
	return r.scan(sql, tenantID, patientID)
}

func (r *PrescriptionProjectionRepository) ListAll(tenantID string) ([]ports.PrescriptionProjection, error) {
	sql := `
		SELECT evidence_id, tenant_id, patient_id, medicamento_generico, dosis,
		       diagnostico, indicaciones, seguimiento, state, created_at, issued_at
		FROM prescription_projections
		WHERE tenant_id = $1 AND state = 'issued'
		ORDER BY created_at DESC`
	return r.scan(sql, tenantID)
}

func (r *PrescriptionProjectionRepository) scan(sql string, args ...any) ([]ports.PrescriptionProjection, error) {
	rows, err := r.pool.Query(context.Background(), sql, args...)
	if err != nil {
		return nil, fmt.Errorf("error al listar proyecciones de recetas: %w", err)
	}
	defer rows.Close()
	var result []ports.PrescriptionProjection
	for rows.Next() {
		var p ports.PrescriptionProjection
		if err := rows.Scan(
			&p.EvidenceID, &p.TenantID, &p.PatientID,
			&p.MedicamentoGenerico, &p.Dosis,
			&p.Diagnostico, &p.Indicaciones, &p.Seguimiento,
			&p.State, &p.CreatedAt, &p.IssuedAt,
		); err != nil {
			return nil, fmt.Errorf("error al escanear proyección de receta: %w", err)
		}
		result = append(result, p)
	}
	return result, nil
}
