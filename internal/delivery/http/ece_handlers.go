package delivery

import (
	"net/http"

	"github.com/Nidael1/VuhmikGO/internal/shaders"
)

// handleECENuevo procesa la captura de un nuevo borrador clínico (draft).
//
// Reglas:
//   - Solo crea registros en estado draft.
//   - No emite ni bloquea registros.
//   - No valida reglas legales o clínicas complejas.
//   - Toda operación es evaluada por el Shader médico.
func handleECENuevo(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		render(w, "ece_draft.html", pageData{
			AppName:     appName(),
			PageTitle:   "Nueva Nota Clínica — Borrador",
			CurrentPath: "/ece/nuevo",
		})

	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "error al procesar formulario", http.StatusBadRequest)
			return
		}

		tenantID := r.Header.Get("X-Tenant-ID")
		actorID := r.Header.Get("X-Actor-ID")
		subjectID := r.FormValue("subject_ref")
		notes := r.FormValue("notes")

		// Validaciones UX — solo presencia y formato
		v := &UXValidator{}
		v.Required("subject_ref", subjectID)
		v.Required("notes", notes)
		v.MaxLength("notes", notes, 2000)

		if !v.Valid() {
			renderUXError(w, r, v.Errors())
			return
		}

		// Evaluación vía Shader — única vía de acceso al Core
		svc := NewShaderService(deliveryDeps.TenantRepo)
		decision := svc.Authorize(tenantID, actorID, shaders.OperationCreate)

		if decision.Result != shaders.DecisionAllow {
			renderShaderDeny(w, r, decision)
			return
		}

		// Operación autorizada — borrador aceptado (persistencia en Sprint posterior)
		render(w, "ece_draft.html", pageData{
			AppName:     appName(),
			PageTitle:   "Borrador registrado",
			CurrentPath: "/ece/nuevo",
		})

	default:
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
	}
}

// handleECEDraftGuardar es el stub del endpoint de guardado de draft.
// La persistencia real se implementa en Issue #35.
// Acepta POST, evalúa via Shader, retorna 202 Accepted.
func handleECEDraftGuardar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "error al procesar formulario", http.StatusBadRequest)
		return
	}
	tenantID := r.Header.Get("X-Tenant-ID")
	actorID := r.Header.Get("X-Actor-ID")
	subjectID := r.FormValue("subject_ref")
	notes := r.FormValue("notes")

	v := &UXValidator{}
	v.Required("subject_ref", subjectID)
	v.Required("notes", notes)
	v.MaxLength("notes", notes, 2000)
	if !v.Valid() {
		renderUXError(w, r, v.Errors())
		return
	}

	svc := NewShaderService(deliveryDeps.TenantRepo)
	decision := svc.Authorize(tenantID, actorID, "create")
	if decision.Result != "allow" {
		renderShaderDeny(w, r, decision)
		return
	}

	// Stub: persistencia completada en Issue #35
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status":"draft","message":"borrador aceptado — persistencia pendiente Issue #35"}`))
}
