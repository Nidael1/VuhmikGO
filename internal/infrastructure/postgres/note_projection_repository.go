package postgres

import (
	"context"
	"fmt"

	"github.com/Nidael1/VuhmikGO/internal/application/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NoteProjectionRepository es el adaptador PostgreSQL para proyecciones
// de notas clínicas (ADR-0022).
type NoteProjectionRepository struct {
	pool *pgxpool.Pool
}

func NewNoteProjectionRepository(pool *pgxpool.Pool) *NoteProjectionRepository {
	return &NoteProjectionRepository{pool: pool}
}

func (r *NoteProjectionRepository) Upsert(p ports.NoteProjection) error {
	sql := `
		INSERT INTO note_projections
			(evidence_id, tenant_id, patient_id, text, state, created_at, issued_at,
			 ta, fc, fr, temp, peso, talla, sao2)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (evidence_id) DO UPDATE SET
			text      = EXCLUDED.text,
			state     = EXCLUDED.state,
			issued_at = EXCLUDED.issued_at,
			ta        = EXCLUDED.ta,
			fc        = EXCLUDED.fc,
			fr        = EXCLUDED.fr,
			temp      = EXCLUDED.temp,
			peso      = EXCLUDED.peso,
			talla     = EXCLUDED.talla,
			sao2      = EXCLUDED.sao2`
	_, err := r.pool.Exec(context.Background(), sql,
		p.EvidenceID, p.TenantID, p.PatientID,
		p.Text, p.State, p.CreatedAt, p.IssuedAt,
		p.TA, p.FC, p.FR, p.Temp, p.Peso, p.Talla, p.SAO2,
	)
	if err != nil {
		return fmt.Errorf("error al guardar proyección de nota: %w", err)
	}
	return nil
}

func (r *NoteProjectionRepository) UpdateState(tenantID, evidenceID, state string) error {
	sql := `
		UPDATE note_projections
		SET state = $1
		WHERE evidence_id = $2 AND tenant_id = $3`
	_, err := r.pool.Exec(context.Background(), sql, state, evidenceID, tenantID)
	if err != nil {
		return fmt.Errorf("error al actualizar estado de proyección de nota: %w", err)
	}
	return nil
}

func (r *NoteProjectionRepository) ListByPatient(tenantID, patientID string) ([]ports.NoteProjection, error) {
	sql := `
		SELECT evidence_id, tenant_id, patient_id, text, state, created_at, issued_at,
		       ta, fc, fr, temp, peso, talla, sao2
		FROM note_projections
		WHERE tenant_id = $1 AND patient_id = $2 AND state = 'issued'
		ORDER BY created_at DESC`
	rows, err := r.pool.Query(context.Background(), sql, tenantID, patientID)
	if err != nil {
		return nil, fmt.Errorf("error al listar proyecciones de notas: %w", err)
	}
	defer rows.Close()
	var result []ports.NoteProjection
	for rows.Next() {
		var p ports.NoteProjection
		if err := rows.Scan(
			&p.EvidenceID, &p.TenantID, &p.PatientID,
			&p.Text, &p.State, &p.CreatedAt, &p.IssuedAt,
			&p.TA, &p.FC, &p.FR, &p.Temp, &p.Peso, &p.Talla, &p.SAO2,
		); err != nil {
			return nil, fmt.Errorf("error al escanear proyección de nota: %w", err)
		}
		result = append(result, p)
	}
	return result, nil
}
