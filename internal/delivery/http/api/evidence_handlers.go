package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Nidael1/VuhmikGO/internal/application"
	"github.com/Nidael1/VuhmikGO/internal/core/evidence"
	"github.com/Nidael1/VuhmikGO/internal/integrity"
	"github.com/Nidael1/VuhmikGO/internal/shaders"
)

// EvidenceItem es el DTO de respuesta para evidencia.
type EvidenceItem struct {
	SubjectRef   string     `json:"subject_ref"`
	Content      string     `json:"content"`
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
		SubjectRef:   e.SubjectRef,
		Content:      e.Content,
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
	items, err := deps.EvidenceRepo.FindAll(tenantID)
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
	e, err := deps.EvidenceRepo.FindByID(tenantID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "evidencia no encontrada")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": toItem(e), "error": nil})
}

// DraftRequest es el payload para crear un borrador.
type DraftRequest struct {
	SubjectRef string `json:"subject_ref"`
	Content    string `json:"content"`
	// Signos vitales opcionales
	TA    string `json:"ta,omitempty"`
	FC    string `json:"fc,omitempty"`
	FR    string `json:"fr,omitempty"`
	Temp  string `json:"temp,omitempty"`
	Peso  string `json:"peso,omitempty"`
	Talla string `json:"talla,omitempty"`
	SAO2  string `json:"sao2,omitempty"`
}

// HandleEvidenceDraft crea un nuevo registro de evidencia en estado draft.
// Delega en NoteService (Handler -> Service -> CapabilityGuard/Shader -> Core).
// Emite automaticamente (ADR-0006) — el frontend nunca debe llamar a
// un endpoint de emit despues de este.
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
	if strings.TrimSpace(req.SubjectRef) == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "subject_id es obligatorio")
		return
	}
	if strings.TrimSpace(req.Content) == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "notes es obligatorio")
		return
	}

	// El frontend envía content como JSON: {"type":"note","text":"..."}
	// Extraemos el texto real para la proyección de lectura (ADR-0022).
	var noteBlob map[string]any
	_ = json.Unmarshal([]byte(req.Content), &noteBlob)
	noteText, _ := noteBlob["text"].(string)

	content := application.NoteContent{
		SubjectRef: req.SubjectRef,
		Text:       noteText,
		TA:         req.TA,
		FC:         req.FC,
		FR:         req.FR,
		Temp:       req.Temp,
		Peso:       req.Peso,
		Talla:      req.Talla,
		SAO2:       req.SAO2,
	}

	e, err := deps.NoteService.Create(tenantID, actorID, req.Content, content)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "NOTE_ERROR", err.Error())
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
	e, err := deps.EvidenceRepo.FindByID(tenantID, id)
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
	e, err := deps.EvidenceRepo.FindByID(tenantID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "evidencia no encontrada")
		return
	}

	// Resolver el modulo real desde el blob y verificar capacidades
	// (ADR-0017) antes de anular — este endpoint es generico y no sabe
	// de antemano a que modulo pertenece el registro.
	moduleID, inner := moduleShaderForBlob(e.Content)
	guard := shaders.NewCapabilityGuard(inner, deps.CapabilityRepo, moduleID, "medico")
	guardCtx := shaders.ShaderContext{
		TenantID:  tenantID,
		Operation: shaders.OperationVoid,
		ActorID:   ActorIDFromContext(r),
	}
	if decision := guard.Evaluate(guardCtx); decision.Result != shaders.DecisionAllow {
		writeError(w, http.StatusUnprocessableEntity, decision.ErrorCode, decision.Reason)
		return
	}

	voided, err := evidence.Void(e, evidence.ReasonCode(req.ReasonCode), time.Now().UTC())
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, mapCoreError(err), err.Error())
		return
	}
	if err := deps.EvidenceRepo.UpdateForVoid(tenantID, voided); err != nil {
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
	orig, err := deps.EvidenceRepo.FindByID(tenantID, id)
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
	// Orden FK crítico: primero Create(nuevo), luego UpdateForVoid(original)
	// para no violar evidence_replaced_by_fk (ADR-0006)
	if err := deps.EvidenceRepo.Create(issued); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al crear reemplazo")
		return
	}
	if err := deps.EvidenceRepo.UpdateForVoid(tenantID, voided); err != nil {
		writeError(w, http.StatusUnprocessableEntity, mapCoreError(err), err.Error())
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

// moduleShaderForBlob lee el campo "type" del blob de contenido y resuelve
// el moduleID + Shader correspondiente (ADR-0016: el Shader interpreta el
// type del blob opaco, nunca el Core). Usado por endpoints genericos de
// evidencia (HandleEvidenceVoid, HandleEvidenceReplace) que no conocen de
// antemano a que modulo pertenece el registro sobre el que operan —
// evita asumir un modulo fijo quien podria no corresponder al contenido
// real (ADR-0017: la compuerta de capacidades debe evaluar el modulo
// real, no un valor hardcodeado).
//
// Tipos reconocidos: allergy, prescription, diagnosis, immunization,
// lab_result. Cualquier otro type (incluyendo "note" o ausente) cae en
// MedicalBasicShader bajo moduleID "note" — hoy es el unico tipo sin
// Shader propio que pasa por estos endpoints genericos.
func moduleShaderForBlob(content string) (moduleID string, s shaders.Shader) {
	var blob struct {
		Type string `json:"type"`
	}
	_ = json.Unmarshal([]byte(content), &blob)

	switch blob.Type {
	case "allergy":
		return "allergy", shaders.NewAllergyShader()
	case "prescription":
		return "prescription", shaders.NewPrescriptionShader()
	case "diagnosis":
		return "diagnosis", shaders.NewDiagnosisShader()
	case "immunization":
		return "immunization", shaders.NewImmunizationShader()
	case "lab_result":
		return "lab_result", shaders.NewLabResultShader()
	default:
		return "note", shaders.NewMedicalBasicShader()
	}
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
	return deps.EvidenceRepo.Update(tenantID, *e)
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
	e, err := deps.EvidenceRepo.FindByID(tenantID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "evidencia no encontrada")
		return
	}

	svc := buildExportShaderForTenant(tenantID) // ADR-0002, issue #208
	ctx := shaders.ShaderContext{
		TenantID:  tenantID,
		ActorID:   actorID,
		Operation: shaders.OperationExport,
		SubjectRef: id,
	}
	data := shaders.ExportData{
		EvidenceID:   e.ID,
		TenantID:     e.TenantID,
		SubjectRef:   e.SubjectRef,
		Content:      e.Content,
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

	// Calcular hash SHA-256 del contenido (ADR-0008)
	issuedStr := ""
	if e.IssuedAt != nil {
		issuedStr = e.IssuedAt.Format(time.RFC3339Nano)
	}
	hashInput := integrity.EvidenceHashInput{
		EvidenceID: e.ID,
		TenantID:   e.TenantID,
		SubjectRef: e.SubjectRef,
		Content:     e.Content,
		State:      string(e.State),
		CreatedAt:  e.CreatedAt.Format(time.RFC3339Nano),
		IssuedAt:   func() *string { if issuedStr == "" { return nil }; return &issuedStr }(),
	}
	contentHash, hashErr := integrity.Hash(hashInput)
	if hashErr != nil {
		contentHash = "sha256:error"
	}

	// Construir respuesta final con hash y contenido completo
	var exportMap map[string]any
	if err := json.Unmarshal(exportBytes, &exportMap); err == nil {
		exportMap["subject_ref"] = e.SubjectRef
		exportMap["content"] = e.Content
		exportMap["hash"] = contentHash
		exportMap["exported_at"] = time.Now().UTC().Format(time.RFC3339Nano)
		if finalBytes, err := json.Marshal(exportMap); err == nil {
			exportBytes = finalBytes
		}
	}

	// Detectar formato segun header Accept (ADR-0007)
	accept := r.Header.Get("Accept")
	if accept == "application/xml" {
		xmlBytes, xmlErr := shaders.GenerateExportXML(data, contentHash)
		if xmlErr == nil {
			w.Header().Set("Content-Type", "application/xml; charset=utf-8")
			w.Header().Set("Content-Disposition", `attachment; filename="export_`+id+`.xml"`)
			w.Header().Set("Cache-Control", "no-store")
			w.WriteHeader(http.StatusOK)
			w.Write(xmlBytes)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="export_`+id+`.json"`)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	w.Write(exportBytes)
}

// buildExportShaderForTenant resuelve el ExportShader por export_shader_key del tenant.
// Fail-closed: key desconocido, vacío o export_none → LegalExportShader como fallback seguro
// para tenants médicos MX (export_shader_key = legal_export por defecto en backfill).
func buildExportShader() shaders.ExportShader {
	return shaders.NewLegalExportShader()
}

func buildExportShaderForTenant(tenantID string) shaders.ExportShader {
	if deps.TenantRepo == nil {
		return shaders.NewLegalExportShader()
	}
	cfg, err := deps.TenantRepo.GetByID(tenantID)
	if err != nil {
		return shaders.NewLegalExportShader() // fail-safe: tenant no encontrado
	}
	registry := shaders.NewExportShaderRegistry()
	if resolved := registry.Resolve(shaders.ExportShaderKey(cfg.ExportShaderKey)); resolved != nil {
		return resolved
	}
	return shaders.NewLegalExportShader() // export_none → fallback seguro
}
// EditRequest es el payload para edicion fluida (ADR-0006).
type EditRequest struct {
	Content string `json:"content"`
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
	actorID := ActorIDFromContext(r)
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

	var req EditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "payload invalido")
		return
	}

	// El frontend envía content como JSON: {"type":"note","text":"..."}
	var noteBlob map[string]any
	_ = json.Unmarshal([]byte(req.Content), &noteBlob)
	noteText, _ := noteBlob["text"].(string)

	e, err := deps.NoteService.Edit(tenantID, actorID, id, req.Content, noteText)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "NOTE_ERROR", err.Error())
		return
	}

	// El medico no ve que hubo un void + replace
	writeJSON(w, http.StatusOK, map[string]any{"data": toItem(e), "error": nil})
}
