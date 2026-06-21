package shaders

// MedicalBasicShader es el Shader médico básico.
//
// Implementa el contrato Shader para operaciones clínicas estándar.
// Aplica validación mínima de frontera: no contiene reglas clínicas
// complejas ni lógica de negocio. No accede al Core directamente.
//
// Perfil: med_basic — CRM clínico base sin requisitos normativos de país.
type MedicalBasicShader struct{}

// shaderErrorCodes para el perfil médico básico.
const (
	ErrShaderContextInvalid  = "ER-SHADER-001"
	ErrShaderOperationDenied = "ER-SHADER-002"
)

// Evaluate evalúa si la operación está permitida en el perfil médico básico.
//
// Reglas:
//   - El contexto debe tener tenant_id, operation y actor_id.
//   - Solo se permiten operaciones declaradas en el catálogo de operaciones.
//   - No se evalúan reglas clínicas; eso corresponde a shaders especializados.
func (s *MedicalBasicShader) Evaluate(ctx ShaderContext) ShaderDecision {
	if err := ctx.Validate(); err != nil {
		return ShaderDecision{
			Result:    DecisionDeny,
			ErrorCode: ErrShaderContextInvalid,
			Reason:    err.Error(),
		}
	}

	switch ctx.Operation {
	case OperationCreate,
		OperationVoid,
		OperationReplace,
		OperationRead,
		OperationExport:
		return ShaderDecision{
			Result: DecisionAllow,
			Reason: "operación permitida en perfil médico básico",
		}
	default:
		return ShaderDecision{
			Result:    DecisionDeny,
			ErrorCode: ErrShaderOperationDenied,
			Reason:    "operación no reconocida en perfil médico básico",
		}
	}
}

// NewMedicalBasicShader retorna una instancia del Shader médico básico.
func NewMedicalBasicShader() Shader {
	return &MedicalBasicShader{}
}
