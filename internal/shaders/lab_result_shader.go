package shaders

import "encoding/json"

// LabResultShader es el Shader para el módulo de resultados de laboratorio (ADR-0015).
type LabResultShader struct{}

// LabResultContent es la estructura del blob para registros tipo lab_result.
// El Core nunca la interpreta; solo el Shader y el Asteroide la conocen.
type LabResultContent struct {
	Type            string `json:"type"`
	Estudio         string `json:"estudio"`
	FechaEstudio    string `json:"fecha_estudio"`
	Resultado       string `json:"resultado,omitempty"`
	Laboratorio     string `json:"laboratorio,omitempty"`
	Unidades        string `json:"unidades,omitempty"`
	ValorReferencia string `json:"valor_referencia,omitempty"`
	Notas           string `json:"notas,omitempty"`
}

func (s *LabResultShader) Evaluate(ctx ShaderContext) ShaderDecision {
	if err := ctx.Validate(); err != nil {
		return ShaderDecision{Result: DecisionDeny, ErrorCode: ErrShaderContextInvalid, Reason: err.Error()}
	}
	switch ctx.Operation {
	case OperationCreate, OperationVoid, OperationReplace, OperationRead, OperationExport:
		return ShaderDecision{Result: DecisionAllow, Reason: "operacion permitida en modulo lab_result"}
	default:
		return ShaderDecision{Result: DecisionDeny, ErrorCode: ErrShaderOperationDenied, Reason: "operacion no reconocida en modulo lab_result"}
	}
}

func ValidateLabResultContent(c LabResultContent) error {
	if c.Estudio == "" {
		return &missingFieldError{"estudio"}
	}
	if c.FechaEstudio == "" {
		return &missingFieldError{"fecha_estudio"}
	}
	return nil
}

func BuildLabResultBlob(c LabResultContent) (string, error) {
	c.Type = "lab_result"
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func ParseLabResultBlob(blob string, c *LabResultContent) error {
	return json.Unmarshal([]byte(blob), c)
}

func NewLabResultShader() Shader { return &LabResultShader{} }
