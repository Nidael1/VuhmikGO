package delivery

import (
	"encoding/json"
	"net/http"
)

// handleECEDraftSave implementa el guardado real del borrador via Shader.
// Reemplaza el stub del Issue #34.
//
// Reglas:
//   - Valida UX antes de llamar al Shader.
//   - El Shader valida y crea el draft en memoria.
//   - Retorna DraftResponse con id y state.
//   - No emite, no bloquea, no registra PHI.
func handleECEDraftSave(w http.ResponseWriter, r *http.Request) {
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
	draft, err := svc.CreateDraft(tenantID, actorID, subjectID)
	if err != nil {
		renderShaderDeny(w, r, decisionFromError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(draft)
}
