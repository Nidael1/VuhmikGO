package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/Nidael1/VuhmikGO/internal/application/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AllergyProjectionRepository es el adaptador PostgreSQL para proyecciones
// de alergias (ADR-0022).
type AllergyProjectionRepository struct {
	pool *pgxpool.Pool
}

func NewAllergyProjectionRepository(pool *pgxpool.Pool) *AllergyProjectionRepository {
	return &AllergyProjectionRepository{pool: pool}
}

func (r *AllergyProjectionRepository) Upsert(p ports.AllergyProjection) error {
	sql := `
		INSERT INTO allergy_projections
			(evidence_id, tenant_id, patient_id, agente, tipo_reaccion,
			 criticidad, certeza, fecha_inicio, notas, state, created_at, issued_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (evidence_id) DO UPDATE SET
			agente        = EXCLUDED.agente,
			tipo_reaccion = EXCLUDED.tipo_reaccion,
			criticidad    = EXCLUDED.criticidad,
			certeza       = EXCLUDED.certeza,
			fecha_inicio  = EXCLUDED.fecha_inicio,
			notas         = EXCLUDED.notas,
			state         = EXCLUDED.state,
			issued_at     = EXCLUDED.issued_at`
	_, err := r.pool.Exec(context.Background(), sql,
		p.EvidenceID, p.TenantID, p.PatientID,
		p.Agente, p.TipoReaccion, p.Criticidad, p.Certeza,
		p.FechaInicio, p.Notas, p.State, p.CreatedAt, p.IssuedAt,
	)
	if err != nil {
		return fmt.Errorf("error al guardar proyección de alergia: %w", err)
	}
	return nil
}

func (r *AllergyProjectionRepository) UpdateState(tenantID, evidenceID, state string) error {
	sql := `
		UPDATE allergy_projections
		SET state = $1
		WHERE evidence_id = $2 AND tenant_id = $3`
	_, err := r.pool.Exec(context.Background(), sql, state, evidenceID, tenantID)
	if err != nil {
		return fmt.Errorf("error al actualizar estado de proyección: %w", err)
	}
	return nil
}

func (r *AllergyProjectionRepository) ListByPatient(tenantID, patientID string) ([]ports.AllergyProjection, error) {
	sql := `
		SELECT evidence_id, tenant_id, patient_id, agente, tipo_reaccion,
		       criticidad, certeza, fecha_inicio, notas, state, created_at, issued_at
		FROM allergy_projections
		WHERE tenant_id = $1 AND patient_id = $2 AND state = 'issued'
		ORDER BY created_at DESC`
	rows, err := r.pool.Query(context.Background(), sql, tenantID, patientID)
	if err != nil {
		return nil, fmt.Errorf("error al listar proyecciones de alergias: %w", err)
	}
	defer rows.Close()
	var result []ports.AllergyProjection
	for rows.Next() {
		var p ports.AllergyProjection
		if err := rows.Scan(
			&p.EvidenceID, &p.TenantID, &p.PatientID,
			&p.Agente, &p.TipoReaccion, &p.Criticidad, &p.Certeza,
			&p.FechaInicio, &p.Notas, &p.State, &p.CreatedAt, &p.IssuedAt,
		); err != nil {
			return nil, fmt.Errorf("error al escanear proyección: %w", err)
		}
		result = append(result, p)
	}
	return result, nil
}

// toProjection convierte una AllergyProjection a tiempo para issued_at
func toIssuedAt(t time.Time) *time.Time {
	return &t
}
