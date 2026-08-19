package shaders

import (
	"fmt"
	"time"
)

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

// ErrShaderDemographic es el error_code para validaciones demográficas (ADR-0030).
const ErrShaderDemographic = "ER-SHADER-010"

// ValidatePatientDemographics valida datos demográficos de un paciente
// conforme a NOM-024-SSA3-2012. Solo valida campos no vacíos (soporta PATCH parcial).
// Retorna nil si los datos son válidos.
func ValidatePatientDemographics(nombre, fechaNacimiento, sexo string) error {
	if fechaNacimiento != "" {
		fn, err := time.Parse("2006-01-02", fechaNacimiento)
		if err != nil {
			return fmt.Errorf("%s: fecha_nacimiento debe tener formato YYYY-MM-DD", ErrShaderDemographic)
		}
		if fn.After(time.Now().UTC().Truncate(24 * time.Hour)) {
			return fmt.Errorf("%s: fecha_nacimiento no puede ser futura", ErrShaderDemographic)
		}
	}
	if sexo != "" && sexo != "M" && sexo != "F" && sexo != "I" {
		return fmt.Errorf("%s: sexo debe ser M, F o I", ErrShaderDemographic)
	}
	return nil
}
