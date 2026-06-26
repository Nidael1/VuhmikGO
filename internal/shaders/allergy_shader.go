package shaders

import "encoding/json"

// AllergyShader es el Shader para el modulo de alergias e intolerancias.
//
// El Core conserva las invariantes (append-only, lifecycle, hash).
// El Shader informa la politica contextual: campos minimos validos.
// Perfil: allergy — ADR-0012.
type AllergyShader struct{}

// AllergyContent es la estructura del blob para registros tipo allergy.
// El Core nunca la interpreta; solo el Shader y el Asteroide la conocen.
type AllergyContent struct {
	Type         string `json:"type"`
	Agente       string `json:"agente"`
	TipoReaccion string `json:"tipo_reaccion"`
	Criticidad   string `json:"criticidad,omitempty"`
	Certeza      string `json:"certeza,omitempty"`
	FechaInicio  string `json:"fecha_inicio,omitempty"`
	Notas        string `json:"notas,omitempty"`
}

// Evaluate evalua si la operacion esta permitida para el modulo allergy.
func (s *AllergyShader) Evaluate(ctx ShaderContext) ShaderDecision {
	if err := ctx.Validate(); err != nil {
		return ShaderDecision{
			Result:    DecisionDeny,
			ErrorCode: ErrShaderContextInvalid,
			Reason:    err.Error(),
		}
	}
	switch ctx.Operation {
	case OperationCreate, OperationVoid, OperationReplace, OperationRead, OperationExport:
		return ShaderDecision{Result: DecisionAllow, Reason: "operacion permitida en modulo allergy"}
	default:
		return ShaderDecision{
			Result:    DecisionDeny,
			ErrorCode: ErrShaderOperationDenied,
			Reason:    "operacion no reconocida en modulo allergy",
		}
	}
}

// ValidateAllergyContent valida los campos minimos del blob de alergia.
func ValidateAllergyContent(content AllergyContent) error {
	if content.Agente == "" {
		return &missingFieldError{"agente"}
	}
	if content.TipoReaccion == "" {
		return &missingFieldError{"tipo_reaccion"}
	}
	return nil
}

// BuildAllergyBlob construye el blob JSON opaco para el Core.
func BuildAllergyBlob(c AllergyContent) (string, error) {
	c.Type = "allergy"
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// NewAllergyShader retorna una instancia del Shader de alergias.
func NewAllergyShader() Shader {
	return &AllergyShader{}
}
