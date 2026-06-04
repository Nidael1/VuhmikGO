package delivery

import "net/http"

// RegisterRoutes registra las rutas del CRM en el ServeMux.
// Solo define navegación base; sin lógica clínica ni reglas del Core.
func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/dashboard", handleDashboard)
	mux.HandleFunc("/pacientes", handlePacientes)
}
