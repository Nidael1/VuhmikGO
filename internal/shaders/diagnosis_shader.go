package shaders

import "encoding/json"

type DiagnosisShader struct{}

type DiagnosisContent struct {
	Type           string `json:"type"`
	Descripcion    string `json:"descripcion"`
	CodigoCIE10    string `json:"codigo_cie10,omitempty"`
	Tipo           string `json:"tipo,omitempty"`
	EstadoProblema string `json:"estado_problema,omitempty"`
	FechaInicio    string `json:"fecha_inicio,omitempty"`
	Notas          string `json:"notas,omitempty"`
}

func (s *DiagnosisShader) Evaluate(ctx ShaderContext) ShaderDecision {
	if err := ctx.Validate(); err != nil {
		return ShaderDecision{Result: DecisionDeny, ErrorCode: ErrShaderContextInvalid, Reason: err.Error()}
	}
	switch ctx.Operation {
	case OperationCreate, OperationVoid, OperationReplace, OperationRead, OperationExport:
		return ShaderDecision{Result: DecisionAllow, Reason: "operacion permitida en modulo diagnosis"}
	default:
		return ShaderDecision{Result: DecisionDeny, ErrorCode: ErrShaderOperationDenied, Reason: "operacion no reconocida en modulo diagnosis"}
	}
}

func ValidateDiagnosisContent(c DiagnosisContent) error {
	if c.Descripcion == "" {
		return &missingFieldError{"descripcion"}
	}
	return nil
}

func BuildDiagnosisBlob(c DiagnosisContent) (string, error) {
	c.Type = "diagnosis"
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func ParseDiagnosisBlob(blob string, c *DiagnosisContent) error {
	return json.Unmarshal([]byte(blob), c)
}

func NewDiagnosisShader() Shader {
	return &DiagnosisShader{}
}
