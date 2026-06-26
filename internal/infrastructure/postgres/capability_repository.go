package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Nidael1/VuhmikGO/internal/application/ports"
)

// CapabilityRepository es el adaptador PostgreSQL para el registro
// de capacidades (ADR-0017). Fail-closed: retorna false ante cualquier
// duda — modulo no encontrado, no publicado, no activo.
type CapabilityRepository struct {
	pool *pgxpool.Pool
}

func NewCapabilityRepository(pool *pgxpool.Pool) *CapabilityRepository {
	return &CapabilityRepository{pool: pool}
}

func (r *CapabilityRepository) IsPublished(moduleID, rubro string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(context.Background(),
		`SELECT EXISTS(
			SELECT 1 FROM modules
			WHERE id = $1 AND rubro = $2 AND publication_status = 'publicado'
		)`, moduleID, rubro).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *CapabilityRepository) IsActive(tenantID, moduleID string) (bool, error) {
	var active bool
	err := r.pool.QueryRow(context.Background(),
		`SELECT COALESCE(
			(SELECT active FROM tenant_capabilities
			 WHERE tenant_id = $1 AND module_id = $2),
			false
		)`, tenantID, moduleID).Scan(&active)
	if err != nil {
		return false, err
	}
	return active, nil
}

func (r *CapabilityRepository) Activate(tenantID, moduleID, plan string, costo float64) error {
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO tenant_capabilities (tenant_id, module_id, active, plan, costo, updated_at)
		 VALUES ($1, $2, true, $3, $4, NOW())
		 ON CONFLICT (tenant_id, module_id)
		 DO UPDATE SET active = true, plan = $3, costo = $4, updated_at = NOW()`,
		tenantID, moduleID, plan, costo)
	return err
}

func (r *CapabilityRepository) Deactivate(tenantID, moduleID string) error {
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO tenant_capabilities (tenant_id, module_id, active, updated_at)
		 VALUES ($1, $2, false, NOW())
		 ON CONFLICT (tenant_id, module_id)
		 DO UPDATE SET active = false, updated_at = NOW()`,
		tenantID, moduleID)
	return err
}

func (r *CapabilityRepository) ListByTenant(tenantID, rubro string) ([]ports.ModuleStatus, error) {
	rows, err := r.pool.Query(context.Background(),
		`SELECT m.id, m.descripcion,
			COALESCE(tc.active, false) AS active,
			COALESCE(tc.plan, '') AS plan,
			COALESCE(tc.costo, 0) AS costo
		 FROM modules m
		 LEFT JOIN tenant_capabilities tc
		   ON tc.module_id = m.id AND tc.tenant_id = $1
		 WHERE m.rubro = $2 AND m.publication_status = 'publicado'
		 ORDER BY m.id`,
		tenantID, rubro)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []ports.ModuleStatus
	for rows.Next() {
		var ms ports.ModuleStatus
		if err := rows.Scan(&ms.ModuleID, &ms.Descripcion,
			&ms.Active, &ms.Plan, &ms.Costo); err != nil {
			return nil, err
		}
		result = append(result, ms)
	}
	return result, nil
}
