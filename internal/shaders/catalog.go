package shaders

// ShaderKey identifica un shader de forma opaca para el Core.
// El Core NO interpreta estos valores (ADR-0002 §2).
type ShaderKey string

const (
	// ShaderGenericCRM — modo base clínico mínimo sin cumplimiento de país.
	// Fallback seguro cuando no hay shader de cumplimiento activo (ADR-0002 §3).
	ShaderGenericCRM ShaderKey = "generic_crm"

	// ShaderMxMedical — cumplimiento normativo médico México (ADR-0002 §4).
	// Extra shader que se suma al clinical shader base med_basic.
	ShaderMxMedical ShaderKey = "mx_medical"

	// ShaderMxTelemedicine2026 — reservado, futuro NO activo (ADR-0002 §5).
	// No implementar. Declarado para bloquear uso accidental.
	ShaderMxTelemedicine2026 ShaderKey = "mx_telemedicine_2026"
)

// ShaderRegistry resuelve el Shader correcto dado un clinical_shader_key.
// El Core trata estos keys como opacos (ADR-0002 §2).
// Fail-closed: key desconocido → MedicalBasicShader (fallback seguro).
type ShaderRegistry struct{}

// NewShaderRegistry retorna una instancia del registry.
func NewShaderRegistry() *ShaderRegistry {
	return &ShaderRegistry{}
}

// Resolve retorna el Shader correspondiente al key indicado.
// Nunca retorna nil. Fail-closed: key inválido → fallback genérico.
func (r *ShaderRegistry) Resolve(key ShaderKey) Shader {
	switch key {
	case ShaderMxMedical:
		return NewMxMedicalShader()
	case ShaderGenericCRM:
		return NewMedicalBasicShader()
	default:
		return NewMedicalBasicShader()
	}
}
