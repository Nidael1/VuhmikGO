package delivery

import "net/http"

// RegisterRoutes registra las rutas del CRM en el ServeMux.
// Solo define navegación base; sin lógica clínica ni reglas del Core.
// Toda ruta sensible queda protegida por TenantContextMiddleware.
func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/dashboard", handleDashboard)
	mux.HandleFunc("/pacientes", handlePacientes)
	mux.HandleFunc("/registros/nuevo", handleNuevoRegistro)
	mux.HandleFunc("/ece/nuevo", handleECENuevo)
	mux.HandleFunc("/ece/draft/guardar", handleECEDraftSave)
	mux.HandleFunc("/ece/emitir", handleECEEmitir)
	mux.HandleFunc("/ece/anular", handleECEVoid)
	mux.HandleFunc("/ece/exportar", handleECEExport)
}

// Handler envuelve el mux con el middleware de contexto de tenant.
// Usar este Handler como entrypoint en cmd/.
func Handler(mux *http.ServeMux) http.Handler {
	return TenantContextMiddleware(mux)
}
