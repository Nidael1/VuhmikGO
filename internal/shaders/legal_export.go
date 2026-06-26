package shaders

import (
	"encoding/json"
	"time"
)

// ExportData contiene los campos exportables de un registro de evidencia.
type ExportData struct {
	EvidenceID   string     `json:"evidence_id"`
	TenantID     string     `json:"tenant_id"`
	SubjectRef   string     `json:"subject_ref"`
	Content      string     `json:"content"`
	State        string     `json:"state"`
	CreatedAt    time.Time  `json:"created_at"`
	IssuedAt     *time.Time `json:"issued_at,omitempty"`
	VoidedAt     *time.Time `json:"voided_at,omitempty"`
	ReplacedByID *string    `json:"replaced_by_id,omitempty"`
}

// ExportShader extiende el contrato Shader con capacidad de export legal.
type ExportShader interface {
	Shader
	GenerateExport(ctx ShaderContext, data ExportData) ([]byte, error)
}

// LegalExportShader implementa el Shader de exportación legal.
//
// Reglas:
//   - Solo acepta OperationExport.
//   - Genera el export en memoria como JSON.
//   - No persiste archivos. El caller usa los bytes y los descarta.
//   - No registra ni expone PHI en logs ni en el output.
type LegalExportShader struct{}

// Evaluate autoriza o deniega la operación de export.
func (s *LegalExportShader) Evaluate(ctx ShaderContext) ShaderDecision {
	if err := ctx.Validate(); err != nil {
		return ShaderDecision{
			Result:    DecisionDeny,
			ErrorCode: ErrShaderContextInvalid,
			Reason:    err.Error(),
		}
	}
	if ctx.Operation != OperationExport {
		return ShaderDecision{
			Result:    DecisionDeny,
			ErrorCode: ErrShaderOperationDenied,
			Reason:    "este shader solo permite operación export",
		}
	}
	return ShaderDecision{
		Result: DecisionAllow,
		Reason: "export legal autorizado",
	}
}

// GenerateExport genera el export en memoria como JSON.
//
// El resultado debe usarse inmediatamente. No debe almacenarse ni persistirse.
// La responsabilidad de descartar los bytes tras uso es del caller.
func (s *LegalExportShader) GenerateExport(ctx ShaderContext, data ExportData) ([]byte, error) {
	decision := s.Evaluate(ctx)
	if decision.Result != DecisionAllow {
		return nil, &exportDeniedError{decision.ErrorCode, decision.Reason}
	}
	return json.Marshal(data)
}

// NewLegalExportShader retorna una instancia del Shader de export legal.
func NewLegalExportShader() ExportShader {
	return &LegalExportShader{}
}

type exportDeniedError struct {
	code   string
	reason string
}

func (e *exportDeniedError) Error() string {
	return "[" + e.code + "] " + e.reason
}
