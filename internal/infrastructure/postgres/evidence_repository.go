package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Nidael1/VuhmikGO/internal/core/evidence"
)

// EvidenceRepository es el adaptador PostgreSQL para evidencia Core.
type EvidenceRepository struct {
	pool *pgxpool.Pool
}

func NewEvidenceRepository(pool *pgxpool.Pool) *EvidenceRepository {
	return &EvidenceRepository{pool: pool}
}

func (r *EvidenceRepository) Create(e evidence.Evidence) error {
	sql := `
		INSERT INTO evidence (id, tenant_id, subject_ref, content, state, created_at, issued_at, voided_at, replaced_by_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.pool.Exec(context.Background(), sql,
		e.ID, e.TenantID, e.SubjectRef, e.Content, string(e.State),
		e.CreatedAt, e.IssuedAt, e.VoidedAt, e.ReplacedByID,
	)
	if err != nil {
		return fmt.Errorf("error al crear evidencia: %w", err)
	}
	return nil
}

func (r *EvidenceRepository) FindByID(tenantID, id string) (evidence.Evidence, error) {
	sql := `
		SELECT id, tenant_id, subject_ref, content, state, created_at, issued_at, voided_at, replaced_by_id
		FROM evidence WHERE id = $1 AND tenant_id = $2`
	row := r.pool.QueryRow(context.Background(), sql, id, tenantID)
	var e evidence.Evidence
	var state string
	err := row.Scan(
		&e.ID, &e.TenantID, &e.SubjectRef, &e.Content, &state,
		&e.CreatedAt, &e.IssuedAt, &e.VoidedAt, &e.ReplacedByID,
	)
	if err != nil {
		return evidence.Evidence{}, fmt.Errorf("registro %s no encontrado: %w", id, err)
	}
	e.State = evidence.State(state)
	return e, nil
}

func (r *EvidenceRepository) Update(tenantID string, e evidence.Evidence) error {
	current, err := r.FindByID(tenantID, e.ID)
	if err != nil {
		return err
	}
	if err := evidence.GuardMutation(current); err != nil {
		return err
	}
	sql := `
		UPDATE evidence
		SET state = $1, issued_at = $2, voided_at = $3, replaced_by_id = $4
		WHERE id = $5 AND tenant_id = $6`
	_, err = r.pool.Exec(context.Background(), sql,
		string(e.State), e.IssuedAt, e.VoidedAt, e.ReplacedByID, e.ID, tenantID,
	)
	if err != nil {
		return fmt.Errorf("error al actualizar evidencia: %w", err)
	}
	return nil
}

// UpdateForVoid permite actualizar el estado a voided sin pasar por GuardMutation.
// Solo para uso interno de void+replace silencioso (ADR-0006).
func (r *EvidenceRepository) UpdateForVoid(tenantID string, e evidence.Evidence) error {
	_, err := r.FindByID(tenantID, e.ID)
	if err != nil {
		return err
	}
	sql := `
		UPDATE evidence
		SET state = $1, voided_at = $2, replaced_by_id = $3
		WHERE id = $4 AND tenant_id = $5`
	_, err = r.pool.Exec(context.Background(), sql,
		string(e.State), e.VoidedAt, e.ReplacedByID, e.ID, tenantID,
	)
	return err
}

func (r *EvidenceRepository) FindAll(tenantID string) ([]evidence.Evidence, error) {
	sql := `
		SELECT id, tenant_id, subject_ref, content, state, created_at, issued_at, voided_at, replaced_by_id
		FROM evidence WHERE tenant_id = $1 ORDER BY created_at DESC`
	rows, err := r.pool.Query(context.Background(), sql, tenantID)
	if err != nil {
		return nil, fmt.Errorf("error al listar evidencias: %w", err)
	}
	defer rows.Close()
	var result []evidence.Evidence
	for rows.Next() {
		var e evidence.Evidence
		var state string
		if err := rows.Scan(
			&e.ID, &e.TenantID, &e.SubjectRef, &e.Content, &state,
			&e.CreatedAt, &e.IssuedAt, &e.VoidedAt, &e.ReplacedByID,
		); err != nil {
			return nil, fmt.Errorf("error al escanear evidencia: %w", err)
		}
		e.State = evidence.State(state)
		result = append(result, e)
	}
	return result, nil
}
