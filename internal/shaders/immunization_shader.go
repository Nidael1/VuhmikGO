package shaders

import "encoding/json"

// ImmunizationShader es el Shader para el modulo de inmunizaciones (ADR-0014).
type ImmunizationShader struct{}

// ImmunizationContent es la estructura del blob para registros tipo immunization.
type ImmunizationContent struct {
	Type            string `json:"type"`
	Vacuna          string `json:"vacuna"`
	FechaAplicacion string `json:"fecha_aplicacion"`
	Lote            string `json:"lote,omitempty"`
	Dosis           string `json:"dosis,omitempty"`
	Via             string `json:"via,omitempty"`
	AplicadaPor     string `json:"aplicada_por,omitempty"`
	Notas           string `json:"notas,omitempty"`
}

func (s *ImmunizationShader) Evaluate(ctx ShaderContext) ShaderDecision {
	if err := ctx.Validate(); err != nil {
		return ShaderDecision{Result: DecisionDeny, ErrorCode: ErrShaderContextInvalid, Reason: err.Error()}
	}
	switch ctx.Operation {
	case OperationCreate, OperationVoid, OperationReplace, OperationRead, OperationExport:
		return ShaderDecision{Result: DecisionAllow, Reason: "operacion permitida en modulo immunization"}
	default:
		return ShaderDecision{Result: DecisionDeny, ErrorCode: ErrShaderOperationDenied, Reason: "operacion no reconocida en modulo immunization"}
	}
}

func ValidateImmunizationContent(c ImmunizationContent) error {
	if c.Vacuna == "" {
		return &missingFieldError{"vacuna"}
	}
	if c.FechaAplicacion == "" {
		return &missingFieldError{"fecha_aplicacion"}
	}
	return nil
}

func BuildImmunizationBlob(c ImmunizationContent) (string, error) {
	c.Type = "immunization"
	b, err := json.Marshal(c)
	if err != nil { return "", err }
	return string(b), nil
}

func ParseImmunizationBlob(blob string, c *ImmunizationContent) error {
	return json.Unmarshal([]byte(blob), c)
}

func NewImmunizationShader() Shader { return &ImmunizationShader{} }
