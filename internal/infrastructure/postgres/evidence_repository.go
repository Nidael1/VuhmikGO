// Package postgres provee el adaptador PostgreSQL para el repositorio de evidencia.
// Usa pgx/v5 con SQL explícito. Sin ORM.
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Nidael1/VuhmikGO/internal/core/evidence"
)

// EvidenceRepository es el adaptador PostgreSQL para evidencia Core.
// Implementa ports.EvidenceRepository.
type EvidenceRepository struct {
	pool *pgxpool.Pool
}

// NewEvidenceRepository retorna un repositorio PostgreSQL usando el pool dado.
func NewEvidenceRepository(pool *pgxpool.Pool) *EvidenceRepository {
	return &EvidenceRepository{pool: pool}
}

// Create inserta un registro Evidence en estado draft.
func (r *EvidenceRepository) Create(e evidence.Evidence) error {
	sql := `
		INSERT INTO evidence (id, tenant_id, state, created_at, issued_at, voided_at, replaced_by_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.pool.Exec(context.Background(), sql,
		e.ID, e.TenantID, string(e.State),
		e.CreatedAt, e.IssuedAt, e.VoidedAt, e.ReplacedByID,
	)
	if err != nil {
		return fmt.Errorf("error al crear evidencia: %w", err)
	}
	return nil
}

// FindByID recupera un registro Evidence por su ID, exigiendo que
// pertenezca a tenantID (Issue #56 — aislamiento multi-tenant).
// Un registro de otro tenant retorna el mismo error que "no encontrado".
func (r *EvidenceRepository) FindByID(tenantID, id string) (evidence.Evidence, error) {
	sql := `
		SELECT id, tenant_id, state, created_at, issued_at, voided_at, replaced_by_id
		FROM evidence WHERE id = $1 AND tenant_id = $2`
	row := r.pool.QueryRow(context.Background(), sql, id, tenantID)

	var e evidence.Evidence
	var state string
	err := row.Scan(
		&e.ID, &e.TenantID, &state,
		&e.CreatedAt, &e.IssuedAt, &e.VoidedAt, &e.ReplacedByID,
	)
	if err != nil {
		return evidence.Evidence{}, fmt.Errorf("registro %s no encontrado: %w", id, err)
	}
	e.State = evidence.State(state)
	return e, nil
}

// Update persiste cambios de estado en un registro existente, exigiendo
// que pertenezca a tenantID (Issue #56 — aislamiento multi-tenant).
// Rechaza si el estado actual en BD es issued o locked (ER-CORE-001).
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
