package delivery

// errorUXCopy mapea error_code a texto UX legible.
//
// Reglas:
//   - No contiene lógica de negocio.
//   - No decide acciones automáticas basadas en el código.
//   - El error_code se muestra siempre, con o sin mapping.
//   - El CRM presenta; el Core/Shaders deciden.
var errorUXCopy = map[string]string{
	"ER-CORE-001":   "Este registro no puede modificarse en su estado actual.",
	"ER-CORE-002":   "La operación solicitada no es válida para el estado actual.",
	"ER-CORE-003":   "Se requiere un motivo para ejecutar esta operación.",
	"ER-CORE-004":   "El registro de reemplazo tiene un identificador inválido.",
	"ER-SHADER-001": "El contexto de la operación está incompleto o es inválido.",
	"ER-SHADER-002": "Esta operación no está permitida en el perfil actual.",
}

// UXCopyFor retorna el texto UX para un error_code.
// Si no existe mapping, retorna el reason original sin modificar.
// El error_code siempre acompaña el mensaje.
func UXCopyFor(errorCode, defaultReason string) string {
	if copy, ok := errorUXCopy[errorCode]; ok {
		return copy
	}
	return defaultReason
}
