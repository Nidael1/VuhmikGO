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
