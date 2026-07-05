package ports

// TenantConfig contiene la configuración del Shader Stack del tenant.
// No expone entidades Core. Solo datos de configuración de políticas.
// Los shader keys son strings opacos (ADR-0002 §2) — la conversión
// a shaders.ShaderKey ocurre en la capa de entrega, no en el puerto.
type TenantConfig struct {
	TenantID          string
	TenantArea        string
	CountryCode       string
	ClinicalShaderKey string // opaco; convertir a shaders.ShaderKey en delivery
	ExportShaderKey   string   // nullable
	ExtraShaderKeys  []string  // 0..N extra shaders activos del tenant (ADR-0025)
}

// TenantRepository define el contrato de acceso a la configuración de tenant.
// Fail-closed: si el tenant no existe, retorna error y el caller deniega.
// La app nunca escribe clinical_shader_key (configuración admin, ADR-0025).
type TenantRepository interface {
	// GetByID retorna la configuración del tenant por su tenant_id.
	// Retorna error si el tenant no existe en la tabla tenants.
	GetByID(tenantID string) (TenantConfig, error)
}
