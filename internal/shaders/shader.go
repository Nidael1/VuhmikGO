// Package shaders define el contrato único de acceso al Core.
//
// Los Shaders son la única vía permitida para interactuar con el Core.
// Ninguna capa exterior (Asteroides, HTTP, UI) accede al Core directamente.
// Los Shaders no contienen lógica clínica ni administrativa; solo contratos.
package shaders

// Operation representa una operación sensible sobre evidencia Core.
type Operation string

const (
	OperationCreate  Operation = "create"
	OperationVoid    Operation = "void"
	OperationReplace Operation = "replace"
	OperationRead    Operation = "read"
	OperationExport  Operation = "export"
)

// ShaderContext es el DTO de entrada al Shader.
// No expone entidades Core. No contiene PHI.
type ShaderContext struct {
	TenantID  string    // obligatorio
	Operation Operation // obligatorio
	SubjectID string    // opcional según operación
	ActorID   string    // obligatorio
	Country   string    // opcional
}

// Validate verifica campos obligatorios del contexto.
func (c ShaderContext) Validate() error {
	if c.TenantID == "" {
		return &missingFieldError{"tenant_id"}
	}
	if c.Operation == "" {
		return &missingFieldError{"operation"}
	}
	if c.ActorID == "" {
		return &missingFieldError{"actor_id"}
	}
	return nil
}

// DecisionResult indica si la operación está permitida o denegada.
type DecisionResult string

const (
	DecisionAllow DecisionResult = "allow"
	DecisionDeny  DecisionResult = "deny"
)

// ShaderDecision es el DTO de salida del Shader.
// No expone entidades Core ni PHI.
type ShaderDecision struct {
	Result    DecisionResult
	ErrorCode string // vacío si allow
	Reason    string // descripción no sensible
}

// Shader es el contrato único de evaluación de operaciones sobre el Core.
// Toda implementación debe ser determinista: mismo contexto, misma decisión.
type Shader interface {
	Evaluate(ctx ShaderContext) ShaderDecision
}

type missingFieldError struct {
	Field string
}

func (e *missingFieldError) Error() string {
	return "campo obligatorio ausente: " + e.Field
}
