package delivery

import "net/http"

// RegisterRoutes registra las rutas del CRM en el ServeMux.
// Solo define navegación base; sin lógica clínica ni reglas del Core.
func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/dashboard", handleDashboard)
	mux.HandleFunc("/pacientes", handlePacientes)
	mux.HandleFunc("/registros/nuevo", handleNuevoRegistro)
	mux.HandleFunc("/ece/nuevo", handleECENuevo)
	mux.HandleFunc("/ece/emitir", handleECEEmitir)
	mux.HandleFunc("/ece/anular", handleECEVoid)
	mux.HandleFunc("/ece/exportar", handleECEExport)
}
