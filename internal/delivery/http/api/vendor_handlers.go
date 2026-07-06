package api

import "net/http"

// HandleVendorList retorna el catálogo de vendedores activos.
// Usado por el panel admin al crear un tenant (ADR-0026).
//
// GET /api/v1/admin/vendors
func HandleVendorList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	if deps.VendorRepo == nil {
		writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"items": []any{}}, "error": nil})
		return
	}
	vendors, err := deps.VendorRepo.ListActive()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al listar vendedores")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"items": vendors}, "error": nil})
}
