package postgres

import (
	"context"
	"fmt"

	"github.com/Nidael1/VuhmikGO/internal/application/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

// VendorRepository implementa ports.VendorRepository sobre PostgreSQL.
// Solo lectura — el catálogo se gestiona vía migraciones/seed (ADR-0026).
type VendorRepository struct {
	db *pgxpool.Pool
}

// NewVendorRepository retorna una instancia del repositorio de vendedores.
func NewVendorRepository(db *pgxpool.Pool) *VendorRepository {
	return &VendorRepository{db: db}
}

// ListActive retorna todos los vendedores activos del catálogo.
func (r *VendorRepository) ListActive() ([]ports.Vendor, error) {
	const q = `
		SELECT vendor_id, name, active, created_at::text
		FROM vendors
		WHERE active = true
		ORDER BY vendor_id`

	rows, err := r.db.Query(context.Background(), q)
	if err != nil {
		return nil, fmt.Errorf("error al listar vendedores: %w", err)
	}
	defer rows.Close()

	var vendors []ports.Vendor
	for rows.Next() {
		var v ports.Vendor
		if err := rows.Scan(&v.VendorID, &v.Name, &v.Active, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("error al escanear vendedor: %w", err)
		}
		vendors = append(vendors, v)
	}
	return vendors, nil
}

// GetByID retorna un vendedor por su vendor_id.
func (r *VendorRepository) GetByID(vendorID string) (ports.Vendor, error) {
	const q = `
		SELECT vendor_id, name, active, created_at::text
		FROM vendors
		WHERE vendor_id = $1
		LIMIT 1`

	var v ports.Vendor
	err := r.db.QueryRow(context.Background(), q, vendorID).
		Scan(&v.VendorID, &v.Name, &v.Active, &v.CreatedAt)
	if err != nil {
		return ports.Vendor{}, fmt.Errorf("vendedor no encontrado: %s", vendorID)
	}
	return v, nil
}
