package delivery

import (
	"net/http"
)

// handleECEExport genera y sirve el export legal bajo demanda.
//
// Reglas:
//   - Solo acepta POST (acción explícita).
//   - El export se genera en memoria y se sirve directamente.
//   - No se guarda ningún archivo en disco.
//   - No se registra PHI en logs.
//   - El archivo es efímero: vive solo en la respuesta HTTP.
func handleECEExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	actorID := r.Header.Get("X-Actor-ID")
	evidenceID := r.FormValue("evidence_id")

	v := &UXValidator{}
	v.Required("evidence_id", evidenceID)
	if !v.Valid() {
		renderUXError(w, r, v.Errors())
		return
	}

	svc := NewShaderService()
	exportBytes, err := svc.Export(tenantID, actorID, evidenceID)
	if err != nil {
		renderShaderDeny(w, r, decisionFromError(err))
		return
	}

	// Servir export directamente — sin persistencia, sin cache
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"export_legal.json\"")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	w.Write(exportBytes)
}
