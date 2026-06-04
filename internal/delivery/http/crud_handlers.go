package delivery

import (
	"net/http"

	"github.com/Nidael1/VuhmikGO/internal/shaders"
)

// handleNuevoRegistro procesa la solicitud de crear un nuevo registro.
// Evalúa la operación vía Shader antes de cualquier acción.
// No accede al Core directamente.
func handleNuevoRegistro(w http.ResponseWriter, r *http.Request) {
	svc := NewShaderService()

	// Evaluación vía Shader — toda operación pasa por aquí
	decision := svc.Authorize(
		r.Header.Get("X-Tenant-ID"),
		r.Header.Get("X-Actor-ID"),
		shaders.OperationCreate,
	)

	if decision.Result != shaders.DecisionAllow {
		http.Error(w, "["+decision.ErrorCode+"] "+decision.Reason, http.StatusForbidden)
		return
	}

	// Operación autorizada — flujo continúa (persistencia en Sprint posterior)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("operación autorizada — registro pendiente de persistencia"))
}
