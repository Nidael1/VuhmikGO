package delivery

import (
	"net/http"

	"github.com/Nidael1/VuhmikGO/internal/shaders"
)

// handleECEEmitir gestiona la acción explícita de emisión de una nota clínica.
//
// Reglas:
//   - La emisión es una acción explícita. No ocurre automáticamente.
//   - Después de emitir, el registro es inmutable (issued → locked).
//   - La edición queda bloqueada en cuanto se emite.
//   - Toda acción es evaluada por el Shader médico.
func handleECEEmitir(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Muestra pantalla de confirmación explícita antes de emitir
		render(w, "ece_emit_confirm.html", pageData{
			AppName:     appName(),
			PageTitle:   "Confirmar Emisión de Nota Clínica",
			CurrentPath: "/ece/emitir",
		})

	case http.MethodPost:
		tenantID := r.Header.Get("X-Tenant-ID")
		actorID := r.Header.Get("X-Actor-ID")

		// Evaluación vía Shader — acción de creación/emisión
		svc := NewShaderService()
		decision := svc.Authorize(tenantID, actorID, shaders.OperationCreate)

		if decision.Result != shaders.DecisionAllow {
			renderShaderDeny(w, r, decision)
			return
		}

		// Emisión autorizada — registro pasa a issued + locked
		// La persistencia real ocurre en el Sprint de repositorio (Sprint posterior)
		render(w, "ece_emit_confirm.html", pageData{
			AppName:      appName(),
			PageTitle:    "Nota emitida y bloqueada",
			CurrentPath:  "/ece/emitir",
			ErrorCode:    "",
			ErrorMessage: "",
		})

	default:
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
	}
}
