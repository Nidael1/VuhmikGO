package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/Nidael1/VuhmikGO/internal/core/evidence"
	"github.com/Nidael1/VuhmikGO/internal/infrastructure/inmemory"
)

// evidenceStore es el repositorio en memoria compartido para demo.
// Sera reemplazado por inyeccion de dependencias en iteracion posterior.
var evidenceStore = inmemory.NewEvidenceRepository()

// EvidenceItem es el DTO de respuesta para evidencia.
type EvidenceItem struct {
	ID           string     `json:"id"`
	TenantID     string     `json:"tenant_id"`
	State        string     `json:"state"`
	CreatedAt    time.Time  `json:"created_at"`
	IssuedAt     *time.Time `json:"issued_at"`
	VoidedAt     *time.Time `json:"voided_at"`
	ReplacedByID *string    `json:"replaced_by_id"`
}

func toItem(e evidence.Evidence) EvidenceItem {
	return EvidenceItem{
		ID:           e.ID,
		TenantID:     e.TenantID,
		State:        string(e.State),
		CreatedAt:    e.CreatedAt,
		IssuedAt:     e.IssuedAt,
		VoidedAt:     e.VoidedAt,
		ReplacedByID: e.ReplacedByID,
	}
}

// HandleEvidenceList retorna todas las evidencias del tenant autenticado.
//
// GET /api/v1/evidence
func HandleEvidenceList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := TenantIDFromContext(r)
	if tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}
	items, err := evidenceStore.FindAll(tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al listar evidencias")
		return
	}
	dtos := make([]EvidenceItem, 0, len(items))
	for _, e := range items {
		dtos = append(dtos, toItem(e))
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"items": dtos}, "error": nil})
}

// HandleEvidenceDetail retorna el detalle de una evidencia por ID.
//
// GET /api/v1/evidence/:id
func HandleEvidenceDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := TenantIDFromContext(r)
	if tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/evidence/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "id requerido")
		return
	}
	e, err := evidenceStore.FindByID(tenantID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "evidencia no encontrada")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": toItem(e), "error": nil})
}
