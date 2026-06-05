// Package observability provee el logger central del sistema.
//
// Reglas absolutas:
//   - Todos los logs usan este paquete. No se permiten log.Printf ni fmt.Println.
//   - Ningún log contiene PHI: sin datos de pacientes, notas clínicas ni identificadores personales.
//   - Los logs son JSON estructurados a stdout (ADR-0001).
//   - Los campos permitidos son técnicos: tenant_id, error_code, operation, request_id.
package observability

import (
	"log/slog"
	"os"
)

// Logger es el logger central del sistema.
// Inicializado una sola vez al arrancar el servidor.
var Logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))

// LogOperation registra el resultado de una operación técnica.
// No acepta PHI. Solo campos técnicos.
func LogOperation(operation, tenantID, errorCode string) {
	if errorCode != "" {
		Logger.Error("operacion fallida",
			"operation", operation,
			"tenant_id", tenantID,
			"error_code", errorCode,
		)
		return
	}
	Logger.Info("operacion completada",
		"operation", operation,
		"tenant_id", tenantID,
	)
}

// LogRequest registra una solicitud HTTP entrante.
// No registra body ni headers con datos personales.
func LogRequest(method, path, tenantID string) {
	Logger.Info("solicitud recibida",
		"method", method,
		"path", path,
		"tenant_id", tenantID,
	)
}

// LogError registra un error técnico con su error_code.
// No registra datos clínicos ni personales.
func LogError(operation, errorCode, message string) {
	Logger.Error("error tecnico",
		"operation", operation,
		"error_code", errorCode,
		"message", message,
	)
}
