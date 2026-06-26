package ports

// CapabilityRepository define el contrato de acceso al registro de
// capacidades (ADR-0017). Es la unica via para verificar si un modulo
// esta publicado y activo para un tenant.
//
// Fail-closed: si el modulo no existe o no esta activo, IsActive = false.
// La app nunca escribe en modules (plano de control = solo migraciones).
type CapabilityRepository interface {
	// IsPublished verifica que el modulo existe en MODULES con
	// publication_status = 'publicado' y el rubro indicado.
	IsPublished(moduleID, rubro string) (bool, error)

	// IsActive verifica que el tenant tiene el modulo activo en
	// TENANT_CAPABILITIES. Retorna false si no existe el registro.
	IsActive(tenantID, moduleID string) (bool, error)

	// Activate activa un modulo para un tenant (panel admin, ADR-0018).
	// Solo puede activar modulos publicados.
	Activate(tenantID, moduleID, plan string, costo float64) error

	// Deactivate desactiva un modulo para un tenant.
	Deactivate(tenantID, moduleID string) error

	// ListByTenant retorna todos los modulos publicados con su estado
	// activo/inactivo para el tenant dado. Para el panel admin.
	ListByTenant(tenantID, rubro string) ([]ModuleStatus, error)
}

// ModuleStatus representa el estado de un modulo para un tenant.
type ModuleStatus struct {
	ModuleID    string
	Descripcion string
	Active      bool
	Plan        string
	Costo       float64
}
