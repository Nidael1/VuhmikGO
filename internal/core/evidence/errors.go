package evidence

import "fmt"

// ErrCore es un error tipado del Core con código de error explícito.
// El catálogo completo de errores se formaliza en un issue posterior.
type ErrCore struct {
	Code    string
	Message string
}

func (e *ErrCore) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// ErrImmutable se emite cuando se intenta mutar un registro
// en estado issued o locked.
// Código: ER-CORE-001
var ErrImmutable = &ErrCore{
	Code:    "ER-CORE-001",
	Message: "el registro no puede ser modificado en su estado actual",
}

// ErrInvalidTransition se emite cuando se intenta una transición
// de estado no declarada en la matriz de lifecycle.
// Código: ER-CORE-002
var ErrInvalidTransition = &ErrCore{
	Code:    "ER-CORE-002",
	Message: "la transición de estado no está permitida",
}

// ErrMissingReasonCode se emite cuando se intenta una operación
// que requiere reason_code y no fue provisto.
// Código: ER-CORE-003
var ErrMissingReasonCode = &ErrCore{
	Code:    "ER-CORE-003",
	Message: "se requiere un reason_code para ejecutar esta operación",
}

// ErrInvalidReplacement se emite cuando el registro de reemplazo
// tiene un ID inválido o igual al original.
// Código: ER-CORE-004
var ErrInvalidReplacement = &ErrCore{
	Code:    "ER-CORE-004",
	Message: "el registro de reemplazo debe tener un ID único y distinto al original",
}
