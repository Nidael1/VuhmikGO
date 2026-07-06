package ports

// Vendor representa un vendedor del catálogo (ADR-0026).
// Provisional fase 1 — atribución comercial de tenants.
type Vendor struct {
	VendorID  string `json:"vendor_id"`
	Name      string `json:"name"`
	Active    bool   `json:"active"`
	CreatedAt string `json:"created_at"`
}

// VendorRepository define el contrato de acceso al catálogo de vendedores.
// Solo lectura desde la aplicación — el catálogo se gestiona vía migraciones/seed.
// ADR-0026: sin lógica comercial adicional.
type VendorRepository interface {
	// ListActive retorna todos los vendedores activos del catálogo.
	ListActive() ([]Vendor, error)
	// GetByID retorna un vendedor por su vendor_id. Error si no existe.
	GetByID(vendorID string) (Vendor, error)
}
