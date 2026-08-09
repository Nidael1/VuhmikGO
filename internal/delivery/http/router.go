package delivery

import (
	"net/http"
	"os"
)

// RegisterRoutes registra las rutas del CRM en el ServeMux.
// Solo define navegación base; sin lógica clínica ni reglas del Core.
// Toda ruta sensible queda protegida por TenantContextMiddleware.
func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/dashboard", handleDashboard)
	mux.HandleFunc("/pacientes", handlePacientes)
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

// RegisterFrontend registra el servidor de archivos estaticos del frontend Vue.
// Sirve ./frontend/dist como raiz. Cualquier ruta no encontrada devuelve
// index.html para que el router de Vue maneje la navegacion del lado del cliente.
// Las rutas /api/ y /ece/ registradas antes en el mux tienen prioridad.
func RegisterFrontend(mux *http.ServeMux) {
	distPath := "frontend/dist"
	if _, err := os.Stat(distPath); os.IsNotExist(err) {
		// En desarrollo local el dist puede no existir; no registrar
		return
	}
	fs := http.FileServer(http.Dir(distPath))
	mux.HandleFunc("/assets/", func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	})
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	})
}

// spaHandler sirve index.html para cualquier ruta no encontrada (SPA fallback).
func SPAHandler(distPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, distPath+"/index.html")
	}
}
