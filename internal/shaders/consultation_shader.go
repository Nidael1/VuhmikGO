package shaders

import "encoding/json"

// ConsultationShader es el Shader para el módulo de consulta médica.
// La consulta agrupa signos vitales, nota clínica y receta opcional.
// ADR-0024.
type ConsultationShader struct{}

// ConsultationContent es la estructura del blob para type: "consultation".
type ConsultationContent struct {
	Type string `json:"type"`
	// Signos vitales (opcionales)
	TA    string `json:"ta,omitempty"`
	FC    string `json:"fc,omitempty"`
	FR    string `json:"fr,omitempty"`
	Temp  string `json:"temp,omitempty"`
	Peso  string `json:"peso,omitempty"`
	Talla string `json:"talla,omitempty"`
	SAO2  string `json:"sao2,omitempty"`
}

func (s *ConsultationShader) Evaluate(ctx ShaderContext) ShaderDecision {
	if err := ctx.Validate(); err != nil {
		return ShaderDecision{
			Result:    DecisionDeny,
			ErrorCode: ErrShaderContextInvalid,
			Reason:    err.Error(),
		}
	}
	switch ctx.Operation {
	case OperationCreate, OperationVoid, OperationReplace, OperationRead, OperationExport:
		return ShaderDecision{Result: DecisionAllow, Reason: "operacion permitida en modulo consultation"}
	default:
		return ShaderDecision{
			Result:    DecisionDeny,
			ErrorCode: ErrShaderOperationDenied,
			Reason:    "operacion no reconocida en modulo consultation",
		}
	}
}

func BuildConsultationBlob(c ConsultationContent) (string, error) {
	c.Type = "consultation"
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func ParseConsultationBlob(blob string, c *ConsultationContent) error {
	return json.Unmarshal([]byte(blob), c)
}

func NewConsultationShader() Shader {
	return &ConsultationShader{}
}
