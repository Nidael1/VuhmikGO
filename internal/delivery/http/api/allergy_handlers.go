package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Nidael1/VuhmikGO/internal/core/evidence"
	"github.com/Nidael1/VuhmikGO/internal/shaders"
)

// --- DTOs ---

// AllergyRequest es el payload para crear una alergia.
type AllergyRequest struct {
	PatientID    string `json:"patient_id"`
	Agente       string `json:"agente"`
	TipoReaccion string `json:"tipo_reaccion"`
	Criticidad   string `json:"criticidad,omitempty"`
	Certeza      string `json:"certeza,omitempty"`
	FechaInicio  string `json:"fecha_inicio,omitempty"`
	Notas        string `json:"notas,omitempty"`
}

// AllergyItem es el DTO de respuesta para una alergia.
type AllergyItem struct {
	ID           string `json:"id"`
	TenantID     string `json:"tenant_id"`
	PatientID    string `json:"patient_id"`
	Agente       string `json:"agente"`
	TipoReaccion string `json:"tipo_reaccion"`
	Criticidad   string `json:"criticidad,omitempty"`
	Certeza      string `json:"certeza,omitempty"`
	FechaInicio  string `json:"fecha_inicio,omitempty"`
	Notas        string `json:"notas,omitempty"`
	State        string `json:"state"`
}

// --- Handlers ---

// HandleAllergyCreate crea una nueva alergia para un paciente.
//
// POST /api/v1/patients/:id/allergies
func HandleAllergyCreate(w http.ResponseWriter, r *http.Request) {
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

	patientID := extractPatientID(r.URL.Path, "/allergies")
	if patientID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "patient_id requerido en la ruta")
		return
	}

	var req AllergyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "payload invalido")
		return
	}
	if strings.TrimSpace(req.Agente) == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "agente es obligatorio")
		return
	}
	if strings.TrimSpace(req.TipoReaccion) == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "tipo_reaccion es obligatorio")
		return
	}

	content := shaders.AllergyContent{
		Agente:       req.Agente,
		TipoReaccion: req.TipoReaccion,
		Criticidad:   req.Criticidad,
		Certeza:      req.Certeza,
		FechaInicio:  req.FechaInicio,
		Notas:        req.Notas,
	}

	e, err := deps.AllergyService.Create(tenantID, actorID, content)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "ALLERGY_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"data":  allergyToItem(e, patientID),
		"error": nil,
	})
}

// HandleAllergyList lista las alergias activas de un paciente.
//
// GET /api/v1/patients/:id/allergies
func HandleAllergyList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := TenantIDFromContext(r)
	if tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}

	patientID := extractPatientID(r.URL.Path, "/allergies")
	if patientID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "patient_id requerido en la ruta")
		return
	}

	items, err := deps.AllergyService.ListByPatient(tenantID, patientID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al listar alergias")
		return
	}

	dtos := make([]AllergyItem, 0, len(items))
	for _, e := range items {
		dtos = append(dtos, allergyToItem(e, patientID))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data":  map[string]any{"items": dtos},
		"error": nil,
	})
}

// HandleAllergyVoid anula una alergia.
//
// POST /api/v1/allergies/:id/void
func HandleAllergyVoid(w http.ResponseWriter, r *http.Request) {
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

	allergyID := strings.TrimPrefix(r.URL.Path, "/api/v1/allergies/")
	allergyID = strings.TrimSuffix(allergyID, "/void")
	if allergyID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "allergy id requerido")
		return
	}

	voided, err := deps.AllergyService.Void(tenantID, actorID, allergyID)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "ALLERGY_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":  toItem(voided),
		"error": nil,
	})
}

// --- helpers ---

// extractPatientID extrae el patient ID de rutas tipo /patients/:id/allergies
func extractPatientID(path, suffix string) string {
	path = strings.TrimPrefix(path, "/api/v1/patients/")
	path = strings.TrimSuffix(path, suffix)
	return path
}

// allergyToItem convierte una evidencia de alergia a DTO de respuesta.
// Parsea el blob JSON para extraer los campos clinicos.
func allergyToItem(e evidence.Evidence, patientID string) AllergyItem {
	var c shaders.AllergyContent
	_ = json.Unmarshal([]byte(e.Content), &c)
	return AllergyItem{
		ID:           e.ID,
		TenantID:     e.TenantID,
		PatientID:    patientID,
		Agente:       c.Agente,
		TipoReaccion: c.TipoReaccion,
		Criticidad:   c.Criticidad,
		Certeza:      c.Certeza,
		FechaInicio:  c.FechaInicio,
		Notas:        c.Notas,
		State:        string(e.State),
	}
}
