package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Nidael1/VuhmikGO/internal/infrastructure/postgres"
	"github.com/Nidael1/VuhmikGO/internal/shaders"
)

// PatientItem es el DTO de respuesta para paciente.
type PatientItem struct {
	CURP            string    `json:"curp"`
	ID              string    `json:"id"`
	TenantID        string    `json:"tenant_id"`
	Nombre          string    `json:"nombre"`
	FechaNacimiento string    `json:"fecha_nacimiento"`
	Sexo            string    `json:"sexo"`
	NumExpediente   string    `json:"num_expediente"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func toPatientItem(p postgres.Patient) PatientItem {
	return PatientItem{
		ID:              p.ID,
		TenantID:        p.TenantID,
		Nombre:          p.Nombre,
		FechaNacimiento: p.FechaNacimiento,
		Sexo:            p.Sexo,
		NumExpediente:   p.NumExpediente,
		CURP: func() string {
			if p.CURP == nil { return "" }
			return *p.CURP
		}(),
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}
}

// PatientRequest es el payload para crear o actualizar un paciente.
type PatientRequest struct {
	CURP            string `json:"curp"`
	Nombre          string `json:"nombre"`
	FechaNacimiento string `json:"fecha_nacimiento"`
	Sexo            string `json:"sexo"`
}

// HandlePatientList retorna todos los pacientes del tenant.
func HandlePatientList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := TenantIDFromContext(r)
	if tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}
	patients, err := deps.PatientRepo.FindAll(tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al listar pacientes")
		return
	}
	items := make([]PatientItem, 0, len(patients))
	for _, p := range patients {
		items = append(items, toPatientItem(p))
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"items": items}, "error": nil})
}

// HandlePatientCreate crea un nuevo paciente.
func HandlePatientCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := TenantIDFromContext(r)
	if tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}
	var req PatientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "payload invalido")
		return
	}
	if strings.TrimSpace(req.Nombre) == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "nombre es obligatorio")
		return
	}
	if strings.TrimSpace(req.FechaNacimiento) == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "fecha_nacimiento es obligatoria")
		return
	}
	if req.Sexo != "M" && req.Sexo != "F" && req.Sexo != "I" {
		writeError(w, http.StatusBadRequest, "INVALID_FIELDS", "sexo debe ser M, F o I")
		return
	}
	expediente, err := deps.PatientRepo.NextExpediente(tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al generar expediente")
		return
	}
	now := time.Now().UTC()
	p := postgres.Patient{
		ID:              "pac-" + tenantID + "-" + now.Format("20060102150405"),
		TenantID:        tenantID,
		Nombre:          strings.ToUpper(strings.TrimSpace(req.Nombre)),
		FechaNacimiento: req.FechaNacimiento,
		Sexo:            req.Sexo,
		NumExpediente:   expediente,
		CURP:            func() *string {
			v := strings.ToUpper(strings.TrimSpace(req.CURP))
			if v == "" { return nil }
			return &v
		}(),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := deps.PatientRepo.Create(p); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al crear paciente")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": toPatientItem(p), "error": nil})
}

// HandlePatientDetail retorna el detalle de un paciente.
func HandlePatientDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := TenantIDFromContext(r)
	if tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/patients/")
	p, err := deps.PatientRepo.FindByID(tenantID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "paciente no encontrado")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": toPatientItem(p), "error": nil})
}

// HandlePatientUpdate actualiza los datos de un paciente.
func HandlePatientUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := TenantIDFromContext(r)
	if tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/patients/")
	id = strings.TrimSuffix(id, "/edit")
	var req PatientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "payload invalido")
		return
	}
	p, err := deps.PatientRepo.FindByID(tenantID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "paciente no encontrado")
		return
	}
	if strings.TrimSpace(req.Nombre) != "" {
		p.Nombre = req.Nombre
	}
	if strings.TrimSpace(req.FechaNacimiento) != "" {
		p.FechaNacimiento = req.FechaNacimiento
	}
	if req.Sexo == "M" || req.Sexo == "F" || req.Sexo == "I" {
		p.Sexo = req.Sexo
	}
	p.UpdatedAt = time.Now().UTC()
	if err := deps.PatientRepo.Update(tenantID, p); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al actualizar paciente")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": toPatientItem(p), "error": nil})
}


// HandlePatientExport genera el export completo del expediente del paciente.
// Incluye datos del paciente, alergias activas y notas emitidas.
// Hash SHA-256 calculado sobre el conjunto completo (ADR-0008).
//
// GET /api/v1/patients/:id/export
func HandlePatientExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := TenantIDFromContext(r)
	actorID := ActorIDFromContext(r)
	if tenantID == "" || actorID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}

	patientID := strings.TrimPrefix(r.URL.Path, "/api/v1/patients/")
	patientID = strings.TrimSuffix(patientID, "/export")
	if patientID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "patient_id requerido")
		return
	}

	// Paciente
	p, err := deps.PatientRepo.FindByID(tenantID, patientID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "paciente no encontrado")
		return
	}

	// Alergias activas
	allergyEvs, err := deps.AllergyService.ListByPatient(tenantID, patientID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al obtener alergias")
		return
	}
	allergyItems := make([]AllergyItem, 0, len(allergyEvs))
	for _, e := range allergyEvs {
		allergyItems = append(allergyItems, allergyToItem(e, patientID))
	}

	// Notas emitidas (no voided, no draft)
	allEvidence, err := deps.EvidenceRepo.FindAll(tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al obtener notas")
		return
	}
	noteItems := make([]EvidenceItem, 0)
	for _, e := range allEvidence {
		if e.SubjectRef == patientID && e.State == "issued" {
			var c map[string]any
			_ = json.Unmarshal([]byte(e.Content), &c)
			if t, ok := c["type"].(string); ok && t == "note" {
				noteItems = append(noteItems, toItem(e))
			}
		}
	}

	exportedAt := time.Now().UTC()

	expediente := map[string]any{
		"paciente":    toPatientItem(p),
		"alergias":    allergyItems,
		"notas":       noteItems,
		"exported_at": exportedAt.Format(time.RFC3339Nano),
		"tenant_id":   tenantID,
	}

	// Hash SHA-256 del expediente completo (ADR-0008)
	expBytes, _ := json.Marshal(expediente)
	h := sha256.Sum256(expBytes)
	expediente["hash"] = "sha256:" + hex.EncodeToString(h[:])

	// Serializar con hash incluido
	finalBytes, err := json.Marshal(expediente)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al serializar expediente")
		return
	}

	// Detectar formato (ADR-0007)
	accept := r.Header.Get("Accept")
	if accept == "application/xml" {
		data := shaders.ExportData{
			EvidenceID: patientID,
			TenantID:   tenantID,
			SubjectRef: patientID,
			Content:    string(expBytes),
			State:      "issued",
			CreatedAt:  exportedAt,
		}
		xmlBytes, xmlErr := shaders.GenerateExportXML(data, "sha256:"+hex.EncodeToString(h[:]))
		if xmlErr == nil {
			w.Header().Set("Content-Type", "application/xml; charset=utf-8")
			w.Header().Set("Content-Disposition", `attachment; filename="expediente_`+p.NumExpediente+`.xml"`)
			w.Header().Set("Cache-Control", "no-store")
			w.WriteHeader(http.StatusOK)
			w.Write(xmlBytes)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="expediente_`+p.NumExpediente+`.json"`)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	w.Write(finalBytes)
}
// HandlePatientExportZIP genera el Audit Package ZIP del expediente del paciente.
// El ZIP se genera en memoria y se sirve directamente. No se persiste.
// Cache-Control: no-store obligatorio (ADR-0027).
//
// GET /api/v1/patients/:id/export/zip
func HandlePatientExportZIP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := TenantIDFromContext(r)
	actorID := ActorIDFromContext(r)
	if tenantID == "" || actorID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}

	patientID := strings.TrimPrefix(r.URL.Path, "/api/v1/patients/")
	patientID = strings.TrimSuffix(patientID, "/export/zip")
	if patientID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "patient_id requerido")
		return
	}

	// Verificar que el paciente existe
	_, err := deps.PatientRepo.FindByID(tenantID, patientID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "paciente no encontrado")
		return
	}

	// Obtener toda la evidencia del tenant para este paciente
	allEvidence, err := deps.EvidenceRepo.FindAll(tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al obtener evidencia")
		return
	}

	// Filtrar evidencia del paciente (issued y voided — todo el historial)
	exportData := make([]shaders.ExportData, 0)
	for _, e := range allEvidence {
		if e.SubjectRef != patientID {
			continue
		}
		ed := shaders.ExportData{
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
		exportData = append(exportData, ed)
	}

	pkg := shaders.AuditPackage{
		PatientID:   patientID,
		TenantID:    tenantID,
		GeneratedAt: time.Now().UTC(),
		Evidence:    exportData,
	}

	zipBytes, err := shaders.BuildAuditPackageZIP(pkg)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al generar paquete de auditoria")
		return
	}

	filename := "auditoria_" + patientID + "_" + time.Now().UTC().Format("20060102150405") + ".zip"
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	w.Write(zipBytes)
}
