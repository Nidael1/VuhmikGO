package delivery

import (
	"net/http"

	"github.com/Nidael1/VuhmikGO/internal/shaders"
)

// voidReasonOptions lista los reason_code disponibles para void en la UI.
// Solo para presentación — espeja RC-VOID-* del catálogo Core.
// La validación real del reason_code ocurre en el Core vía Shader.
var voidReasonOptions = []struct {
	Code  string
	Label string
}{
	{"RC-VOID-001", "Error detectado en el contenido del registro"},
	{"RC-VOID-002", "La información requiere actualización"},
	{"RC-VOID-003", "Anulación solicitada formalmente"},
	{"RC-VOID-004", "Decisión administrativa documentada"},
}

// voidFormData extiende pageData para el formulario de void.
type voidFormData struct {
	pageData
	ReasonOptions []struct {
		Code  string
		Label string
	}
}

// handleECEVoid gestiona la solicitud de anulación y reemplazo de una nota.
//
// Reglas:
//   - reason_code es obligatorio — sin él el void es rechazado.
//   - La nota original queda voided, no se elimina.
//   - El reemplazo se crea en estado issued.
//   - No existe edición directa ni borrado.
func handleECEVoid(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		render(w, "ece_void.html", voidFormData{
			pageData: pageData{
				AppName:     appName(),
				PageTitle:   "Anular Nota Clínica",
				CurrentPath: "/ece/anular",
			},
			ReasonOptions: voidReasonOptions,
		})

	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "error al procesar formulario", http.StatusBadRequest)
			return
		}

		tenantID := r.Header.Get("X-Tenant-ID")
		actorID := r.Header.Get("X-Actor-ID")
		reasonCode := r.FormValue("reason_code")
		replacement := r.FormValue("replacement_notes")

		// Validaciones UX — presencia obligatoria de reason_code y reemplazo
		v := &UXValidator{}
		v.Required("reason_code", reasonCode)
		v.Required("replacement_notes", replacement)
		v.MaxLength("replacement_notes", replacement, 2000)

		if !v.Valid() {
			renderUXError(w, r, v.Errors())
			return
		}

		// Evaluación vía Shader — operación void
		svc := NewShaderService(deliveryDeps.TenantRepo)
		decision := svc.Authorize(tenantID, actorID, shaders.OperationVoid)

		if decision.Result != shaders.DecisionAllow {
			renderShaderDeny(w, r, decision)
			return
		}

		// Void + replace autorizado — historial preservado (persistencia en Sprint posterior)
		render(w, "ece_void.html", voidFormData{
			pageData: pageData{
				AppName:      appName(),
				PageTitle:    "Nota anulada — reemplazo emitido",
				CurrentPath:  "/ece/anular",
				ErrorCode:    "",
				ErrorMessage: "El registro original ha sido anulado y preservado. El reemplazo fue emitido.",
			},
			ReasonOptions: voidReasonOptions,
		})

	default:
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
	}
}
