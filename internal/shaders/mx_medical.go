package shaders

// MxMedicalShader es el extra shader de cumplimiento normativo para México.
//
// Aplica reglas NOM-024-SSA3-2012 que antes vivían en prescription_handlers.go.
// Se suma al clinical shader base (med_basic) como extra shader de cumplimiento
// de país (ADR-0002, rol policy/extra).
//
// Perfil: mx_medical — cumplimiento médico México.
// Rol: extra (policy) — no reemplaza al clinical shader; se suma.
type MxMedicalShader struct{}

// Evaluate evalúa cumplimiento normativo NOM-024 para operaciones médicas MX.
// Puede endurecer lo que med_basic permite.
// NUNCA puede permitir lo que el Core prohíbe (ADR-0002).
func (s *MxMedicalShader) Evaluate(ctx ShaderContext) ShaderDecision {
	if err := ctx.Validate(); err != nil {
		return ShaderDecision{
			Result:    DecisionDeny,
			ErrorCode: ErrShaderContextInvalid,
			Reason:    err.Error(),
		}
	}
	switch ctx.Operation {
	case OperationCreate, OperationVoid, OperationReplace,
		OperationRead, OperationExport:
		return ShaderDecision{
			Result: DecisionAllow,
			Reason: "operación permitida bajo cumplimiento NOM-024 México",
		}
	default:
		return ShaderDecision{
			Result:    DecisionDeny,
			ErrorCode: ErrShaderOperationDenied,
			Reason:    "operación no reconocida en mx_medical",
		}
	}
}

// MxMedicalProfile representa los campos del perfil médico requeridos por NOM-024.
// El handler los obtiene del repositorio y los pasa a ValidateMxMedicalProfile.
type MxMedicalProfile struct {
	CedulaProfesional string
	Especialidad      string
}

// ValidateMxMedicalProfile valida que el perfil del médico cumple NOM-024.
// Obligatorio antes de emitir cualquier evidencia clínica bajo cumplimiento MX.
// cedula_profesional y especialidad son campos no negociables (NOM-024-SSA3-2012).
func ValidateMxMedicalProfile(p MxMedicalProfile) error {
	if p.CedulaProfesional == "" {
		return &missingFieldError{"cedula_profesional (NOM-024)"}
	}
	if p.Especialidad == "" {
		return &missingFieldError{"especialidad (NOM-024)"}
	}
	return nil
}

// NewMxMedicalShader retorna una instancia del shader de cumplimiento MX.
func NewMxMedicalShader() Shader {
	return &MxMedicalShader{}
}
