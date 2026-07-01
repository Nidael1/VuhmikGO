package postgres

import (
	"context"
	"fmt"

	"github.com/Nidael1/VuhmikGO/internal/application/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ConsultationProjectionRepository struct {
	pool *pgxpool.Pool
}

func NewConsultationProjectionRepository(pool *pgxpool.Pool) *ConsultationProjectionRepository {
	return &ConsultationProjectionRepository{pool: pool}
}

func (r *ConsultationProjectionRepository) Upsert(p ports.ConsultationProjection) error {
	sql := `
		INSERT INTO consultation_projections
			(evidence_id, tenant_id, patient_id, ta, fc, fr, temp, peso, talla, sao2,
			state, created_at, issued_at, tiene_receta)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (evidence_id) DO UPDATE SET
			ta           = EXCLUDED.ta,
			fc           = EXCLUDED.fc,
			fr           = EXCLUDED.fr,
			temp         = EXCLUDED.temp,
			peso         = EXCLUDED.peso,
			talla        = EXCLUDED.talla,
			sao2         = EXCLUDED.sao2,
			state        = EXCLUDED.state,
			issued_at    = EXCLUDED.issued_at,
			tiene_receta = EXCLUDED.tiene_receta`
	_, err := r.pool.Exec(context.Background(), sql,
		p.EvidenceID, p.TenantID, p.PatientID,
		p.TA, p.FC, p.FR, p.Temp, p.Peso, p.Talla, p.SAO2,
		p.State, p.CreatedAt, p.IssuedAt, p.TieneReceta,
	)
	return err
}

func (r *ConsultationProjectionRepository) UpdateState(tenantID, evidenceID, state string) error {
	sql := `UPDATE consultation_projections SET state = $1 WHERE evidence_id = $2 AND tenant_id = $3`
	_, err := r.pool.Exec(context.Background(), sql, state, evidenceID, tenantID)
	return err
}

func (r *ConsultationProjectionRepository) FindByID(tenantID, evidenceID string) (ports.ConsultationProjection, error) {
	sql := `
		SELECT evidence_id, tenant_id, patient_id, ta, fc, fr, temp, peso, talla, sao2,
		       state, created_at, issued_at, tiene_receta
		FROM consultation_projections
		WHERE evidence_id = $1 AND tenant_id = $2`
	var p ports.ConsultationProjection
	err := r.pool.QueryRow(context.Background(), sql, evidenceID, tenantID).Scan(
		&p.EvidenceID, &p.TenantID, &p.PatientID,
		&p.TA, &p.FC, &p.FR, &p.Temp, &p.Peso, &p.Talla, &p.SAO2,
		&p.State, &p.CreatedAt, &p.IssuedAt, &p.TieneReceta,
	)
	if err != nil {
		return ports.ConsultationProjection{}, fmt.Errorf("consulta no encontrada: %w", err)
	}
	return p, nil
}

func (r *ConsultationProjectionRepository) ListByPatient(tenantID, patientID string) ([]ports.ConsultationProjection, error) {
	sql := `
		SELECT evidence_id, tenant_id, patient_id, ta, fc, fr, temp, peso, talla, sao2,
		       state, created_at, issued_at, tiene_receta
		FROM consultation_projections
		WHERE tenant_id = $1 AND patient_id = $2 AND state = 'issued'
		ORDER BY created_at DESC`
	return r.scanAll(sql, tenantID, patientID)
}

func (r *ConsultationProjectionRepository) ListAll(tenantID string) ([]ports.ConsultationProjection, error) {
	sql := `
		SELECT evidence_id, tenant_id, patient_id, ta, fc, fr, temp, peso, talla, sao2,
		       state, created_at, issued_at, tiene_receta
		FROM consultation_projections
		WHERE tenant_id = $1 AND state = 'issued'
		ORDER BY created_at DESC`
	return r.scanAll(sql, tenantID)
}

func (r *ConsultationProjectionRepository) scanAll(sql string, args ...any) ([]ports.ConsultationProjection, error) {
	rows, err := r.pool.Query(context.Background(), sql, args...)
	if err != nil {
		return nil, fmt.Errorf("error al listar consultas: %w", err)
	}
	defer rows.Close()
	var result []ports.ConsultationProjection
	for rows.Next() {
		var p ports.ConsultationProjection
		if err := rows.Scan(
			&p.EvidenceID, &p.TenantID, &p.PatientID,
			&p.TA, &p.FC, &p.FR, &p.Temp, &p.Peso, &p.Talla, &p.SAO2,
			&p.State, &p.CreatedAt, &p.IssuedAt, &p.TieneReceta,
		); err != nil {
			return nil, fmt.Errorf("error al escanear consulta: %w", err)
		}
		result = append(result, p)
	}
	return result, nil
}
