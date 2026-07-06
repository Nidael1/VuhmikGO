package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Nidael1/VuhmikGO/internal/application/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TenantRepository implementa ports.TenantRepository sobre PostgreSQL.
// Solo lectura — la configuración del tenant se gestiona desde el panel admin.
type TenantRepository struct {
	db *pgxpool.Pool
}

// NewTenantRepository retorna una instancia del repositorio de tenant.
func NewTenantRepository(db *pgxpool.Pool) *TenantRepository {
	return &TenantRepository{db: db}
}

// GetByID retorna la configuración del tenant por tenant_id.
// Fail-closed: si el tenant no existe, retorna error.
func (r *TenantRepository) GetByID(tenantID string) (ports.TenantConfig, error) {
	const q = `
		SELECT tenant_id, tenant_area, country_code,
		       clinical_shader_key, export_shader_key
		FROM tenants
		WHERE tenant_id = $1
		LIMIT 1`

	var cfg ports.TenantConfig
	var exportKey sql.NullString

	err := r.db.QueryRow(context.Background(), q, tenantID).Scan(
		&cfg.TenantID,
		&cfg.TenantArea,
		&cfg.CountryCode,
		&cfg.ClinicalShaderKey,
		&exportKey,
	)
	if err != nil {
		return ports.TenantConfig{},
			fmt.Errorf("tenant no encontrado: %s", tenantID)
	}

	if exportKey.Valid {
		cfg.ExportShaderKey = exportKey.String
	}

	// Leer extra shaders activos del tenant (ADR-0025, issue #206).
	// Fail-closed: si falla la query, retorna cfg sin extras (no deniega — el clinical shader sigue activo).
	const qExtra = `
		SELECT shader_key
		FROM tenant_extra_shaders
		WHERE tenant_id = $1 AND active = true
		ORDER BY shader_key`

	rows, err := r.db.Query(context.Background(), qExtra, tenantID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var key string
			if err := rows.Scan(&key); err == nil {
				cfg.ExtraShaderKeys = append(cfg.ExtraShaderKeys, key)
			}
		}
	}

	return cfg, nil
}

// SetVendorRef asigna el vendor_ref a un tenant existente (ADR-0026, issue #220).
// Provisional fase 1 — sin lógica comercial adicional.
func (r *TenantRepository) SetVendorRef(tenantID, vendorID string) error {
	const q = `UPDATE tenants SET vendor_ref = $1, updated_at = NOW() WHERE tenant_id = $2`
	_, err := r.db.Exec(context.Background(), q, vendorID, tenantID)
	if err != nil {
		return fmt.Errorf("error al asignar vendor_ref: %w", err)
	}
	return nil
}
