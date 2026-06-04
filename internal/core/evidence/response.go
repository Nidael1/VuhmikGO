package evidence

// CoreErrorResponse es la representación estructurada de un error del Core.
//
// Formato estable y consistente para todas las respuestas de error.
// No expone PHI ni datos internos del sistema.
// El campo error_code es siempre uno del catálogo canónico.
type CoreErrorResponse struct {
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
}

// ToResponse convierte un ErrCore a su representación estructurada.
// Usar este método garantiza formato uniforme en todas las respuestas.
func (e *ErrCore) ToResponse() CoreErrorResponse {
	return CoreErrorResponse{
		ErrorCode: e.Code,
		Message:   e.Message,
	}
}

// ExtractErrorCode retorna el error_code de un error del Core.
// Retorna cadena vacía si el error no es un ErrCore.
//
// Usar para inspeccionar errores sin exponer la estructura interna
// a capas superiores (Shaders, Asteroides, HTTP).
func ExtractErrorCode(err error) string {
	if e, ok := err.(*ErrCore); ok {
		return e.Code
	}
	return ""
}
