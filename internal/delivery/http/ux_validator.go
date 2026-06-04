package delivery

import (
	"fmt"
	"strings"
	"time"
)

// UXValidationError representa un error de validación a nivel UX.
// No contiene códigos del Core ni lógica de negocio.
type UXValidationError struct {
	Field   string
	Message string
}

func (e *UXValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// UXValidator valida entradas de formulario a nivel UX.
//
// Reglas:
//   - Solo valida presencia (required) y formato.
//   - No valida reglas de negocio ni reglas del Core.
//   - El Core es la única fuente de verdad sobre validez clínica.
type UXValidator struct {
	errors []UXValidationError
}

// Required valida que un campo no esté vacío.
func (v *UXValidator) Required(field, value string) *UXValidator {
	if strings.TrimSpace(value) == "" {
		v.errors = append(v.errors, UXValidationError{
			Field:   field,
			Message: "este campo es obligatorio",
		})
	}
	return v
}

// MaxLength valida que un campo no exceda el máximo de caracteres.
func (v *UXValidator) MaxLength(field, value string, max int) *UXValidator {
	if len([]rune(value)) > max {
		v.errors = append(v.errors, UXValidationError{
			Field:   field,
			Message: fmt.Sprintf("máximo %d caracteres permitidos", max),
		})
	}
	return v
}

// DateFormat valida que un campo tenga formato de fecha válido (YYYY-MM-DD).
func (v *UXValidator) DateFormat(field, value string) *UXValidator {
	if strings.TrimSpace(value) == "" {
		return v
	}
	if _, err := time.Parse("2006-01-02", value); err != nil {
		v.errors = append(v.errors, UXValidationError{
			Field:   field,
			Message: "formato de fecha inválido — use YYYY-MM-DD",
		})
	}
	return v
}

// Valid retorna true si no hay errores de validación UX.
func (v *UXValidator) Valid() bool {
	return len(v.errors) == 0
}

// Errors retorna todos los errores de validación UX.
func (v *UXValidator) Errors() []UXValidationError {
	return v.errors
}
