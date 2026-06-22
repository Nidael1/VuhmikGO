package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Nidael1/VuhmikGO/internal/core/evidence"
	"github.com/Nidael1/VuhmikGO/internal/infrastructure/inmemory"
	"github.com/Nidael1/VuhmikGO/internal/shaders"
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

// DraftRequest es el payload para crear un borrador.
type DraftRequest struct {
	SubjectID string `json:"subject_id"`
	Notes     string `json:"notes"`
}

// HandleEvidenceDraft crea un nuevo registro de evidencia en estado draft.
//
// POST /api/v1/evidence/draft
func HandleEvidenceDraft(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := TenantIDFromContext(r)
	actorID := ActorIDFromContext(r)
	if tenantID == "" || actorID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}
	var req DraftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "payload invalido")
		return
	}
	if strings.TrimSpace(req.SubjectID) == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "subject_id es obligatorio")
		return
	}
	if strings.TrimSpace(req.Notes) == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "notes es obligatorio")
		return
	}
	id := "ev-" + tenantID + "-" + req.SubjectID + "-" + time.Now().Format("20060102150405")
	e := evidence.Evidence{
		ID:        id,
		TenantID:  tenantID,
		State:     evidence.StateDraft,
		CreatedAt: time.Now().UTC(),
	}
	if err := evidenceStore.Create(e); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al crear borrador")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": toItem(e), "error": nil})
}

// VoidRequest es el payload para anular una evidencia.
type VoidRequest struct {
	ReasonCode string `json:"reason_code"`
}

// ReplaceRequest es el payload para reemplazar una evidencia.
type ReplaceRequest struct {
	ReasonCode    string `json:"reason_code"`
	ReplacementID string `json:"replacement_id"`
}

// HandleEvidenceEmit emite una evidencia draft (draft -> issued -> locked).
//
// POST /api/v1/evidence/:id/emit
func HandleEvidenceEmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := TenantIDFromContext(r)
	if tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}
	id := extractID(r.URL.Path, "/emit")
	if id == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "id requerido")
		return
	}
	e, err := evidenceStore.FindByID(tenantID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "evidencia no encontrada")
		return
	}
	now := time.Now().UTC()
	if err := transitionTo(&e, evidence.StateIssued, tenantID, now); err != nil {
		writeError(w, http.StatusUnprocessableEntity, mapCoreError(err), err.Error())
		return
	}
	if err := transitionTo(&e, evidence.StateLocked, tenantID, now); err != nil {
		writeError(w, http.StatusUnprocessableEntity, mapCoreError(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": toItem(e), "error": nil})
}

// HandleEvidenceVoid anula una evidencia con reason_code obligatorio.
//
// POST /api/v1/evidence/:id/void
func HandleEvidenceVoid(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := TenantIDFromContext(r)
	if tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}
	id := extractID(r.URL.Path, "/void")
	if id == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "id requerido")
		return
	}
	var req VoidRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ReasonCode == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "reason_code es obligatorio")
		return
	}
	e, err := evidenceStore.FindByID(tenantID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "evidencia no encontrada")
		return
	}
	voided, err := evidence.Void(e, evidence.ReasonCode(req.ReasonCode), time.Now().UTC())
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, mapCoreError(err), err.Error())
		return
	}
	if err := evidenceStore.Update(tenantID, voided); err != nil {
		writeError(w, http.StatusUnprocessableEntity, mapCoreError(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": toItem(voided), "error": nil})
}

// HandleEvidenceReplace anula el original y crea el reemplazo.
//
// POST /api/v1/evidence/:id/replace
func HandleEvidenceReplace(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := TenantIDFromContext(r)
	if tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}
	id := extractID(r.URL.Path, "/replace")
	if id == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "id requerido")
		return
	}
	var req ReplaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ReasonCode == "" || req.ReplacementID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "reason_code y replacement_id son obligatorios")
		return
	}
	orig, err := evidenceStore.FindByID(tenantID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "evidencia no encontrada")
		return
	}
	repl := evidence.Evidence{
		ID:        req.ReplacementID,
		TenantID:  tenantID,
		State:     evidence.StateDraft,
		CreatedAt: time.Now().UTC(),
	}
	voided, issued, err := evidence.Replace(orig, repl, evidence.ReasonCode(req.ReasonCode), time.Now().UTC())
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, mapCoreError(err), err.Error())
		return
	}
	if err := evidenceStore.Update(tenantID, voided); err != nil {
		writeError(w, http.StatusUnprocessableEntity, mapCoreError(err), err.Error())
		return
	}
	if err := evidenceStore.Create(issued); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al crear reemplazo")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data":  map[string]any{"voided": toItem(voided), "replacement": toItem(issued)},
		"error": nil,
	})
}

// extractID extrae el ID de la URL removiendo el sufijo de accion.
func extractID(path, suffix string) string {
	path = strings.TrimPrefix(path, "/api/v1/evidence/")
	return strings.TrimSuffix(path, suffix)
}

// transitionTo aplica una transicion de estado y persiste.
func transitionTo(e *evidence.Evidence, next evidence.State, tenantID string, now time.Time) error {
	if err := evidence.GuardTransition(e.State, next); err != nil {
		return err
	}
	e.State = next
	if next == evidence.StateIssued {
		e.IssuedAt = &now
	}
	return evidenceStore.Update(tenantID, *e)
}

// mapCoreError mapea errores del Core a codigos de la API.
func mapCoreError(err error) string {
	switch evidence.ExtractErrorCode(err) {
	case "ER-CORE-001":
		return "EVIDENCE_IMMUTABLE"
	case "ER-CORE-002":
		return "EVIDENCE_INVALID_TRANSITION"
	case "ER-CORE-003":
		return "EVIDENCE_MISSING_REASON"
	case "ER-CORE-004":
		return "EVIDENCE_INVALID_REPLACE"
	default:
		return "INTERNAL_ERROR"
	}
}

// HandleEvidenceExport genera el export legal efimero de una evidencia.
//
// POST /api/v1/evidence/:id/export
// El archivo se genera en memoria y se sirve directamente.
// No se persiste ningun archivo. Cache-Control: no-store.
func HandleEvidenceExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := TenantIDFromContext(r)
	actorID := ActorIDFromContext(r)
	if tenantID == "" || actorID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}
	id := extractID(r.URL.Path, "/export")
	if id == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "id requerido")
		return
	}
	e, err := evidenceStore.FindByID(tenantID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "evidencia no encontrada")
		return
	}

	svc := buildExportShader()
	ctx := shaders.ShaderContext{
		TenantID:  tenantID,
		ActorID:   actorID,
		Operation: shaders.OperationExport,
		SubjectID: id,
	}
	data := shaders.ExportData{
		EvidenceID:   e.ID,
		TenantID:     e.TenantID,
		State:        string(e.State),
		CreatedAt:    e.CreatedAt,
		IssuedAt:     e.IssuedAt,
		VoidedAt:     e.VoidedAt,
		ReplacedByID: e.ReplacedByID,
	}
	exportBytes, err := svc.GenerateExport(ctx, data)
	if err != nil {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "export no autorizado")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"export_"+id+".json\"")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	w.Write(exportBytes)
}

func buildExportShader() shaders.ExportShader {
	return shaders.NewLegalExportShader()
}

// HandleEvidenceEdit implementa la edicion fluida segun ADR-0006.
// El medico percibe que edita. Internamente: void + replace silencioso.
// El Core no cambia. La inmutabilidad es total.
//
// PUT /api/v1/evidence/:id
func HandleEvidenceEdit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := TenantIDFromContext(r)
	if tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/evidence/")
	id = strings.TrimSuffix(id, "/edit")
	if id == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "id requerido")
		return
	}

	orig, err := evidenceStore.FindByID(tenantID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "evidencia no encontrada")
		return
	}

	// Si esta en draft, solo actualiza el registro existente (reemite)
	// Si esta emitida/bloqueada, ejecuta void + replace silencioso
	now := time.Now().UTC()
	newID := id + "-v" + now.Format("20060102150405")

	if orig.State == evidence.StateDraft {
		// Solo re-emite el draft existente
		if err := transitionTo(&orig, evidence.StateIssued, tenantID, now); err != nil {
			writeError(w, http.StatusUnprocessableEntity, mapCoreError(err), err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": toItem(orig), "error": nil})
		return
	}

	// Void + replace silencioso (ADR-0006)
	// El medico no ve este proceso — solo ve el resultado final
	repl := evidence.Evidence{
		ID:        newID,
		TenantID:  tenantID,
		State:     evidence.StateDraft,
		CreatedAt: now,
	}

	voided, issued, err := evidence.Replace(
		orig, repl,
		evidence.RCReplaceUpdate, // RC-REPLACE-002 automatico
		now,
	)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, mapCoreError(err), err.Error())
		return
	}

	if err := evidenceStore.Update(tenantID, voided); err != nil {
		writeError(w, http.StatusUnprocessableEntity, mapCoreError(err), err.Error())
		return
	}
	if err := evidenceStore.Create(issued); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al crear version")
		return
	}

	// Retorna el nuevo registro como si fuera el mismo
	// El medico no ve que hubo un void + replace
	writeJSON(w, http.StatusOK, map[string]any{"data": toItem(issued), "error": nil})
}
