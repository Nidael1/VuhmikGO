package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Nidael1/VuhmikGO/internal/core/evidence"
	"github.com/Nidael1/VuhmikGO/internal/shaders"
)

type LabResultRequest struct {
	PatientID       string `json:"patient_id"`
	Estudio         string `json:"estudio"`
	FechaEstudio    string `json:"fecha_estudio"`
	Resultado       string `json:"resultado,omitempty"`
	Laboratorio     string `json:"laboratorio,omitempty"`
	Unidades        string `json:"unidades,omitempty"`
	ValorReferencia string `json:"valor_referencia,omitempty"`
	Notas           string `json:"notas,omitempty"`
}

type LabResultItem struct {
	ID              string `json:"id"`
	TenantID        string `json:"tenant_id"`
	PatientID       string `json:"patient_id"`
	Estudio         string `json:"estudio"`
	FechaEstudio    string `json:"fecha_estudio"`
	Resultado       string `json:"resultado,omitempty"`
	Laboratorio     string `json:"laboratorio,omitempty"`
	Unidades        string `json:"unidades,omitempty"`
	ValorReferencia string `json:"valor_referencia,omitempty"`
	Notas           string `json:"notas,omitempty"`
	State           string `json:"state"`
}

func toLabResultItem(e evidence.Evidence, patientID string) LabResultItem {
	var c shaders.LabResultContent
	_ = shaders.ParseLabResultBlob(e.Content, &c)
	return LabResultItem{
		ID: e.ID, TenantID: e.TenantID, PatientID: patientID,
		Estudio: c.Estudio, FechaEstudio: c.FechaEstudio,
		Resultado: c.Resultado, Laboratorio: c.Laboratorio,
		Unidades: c.Unidades, ValorReferencia: c.ValorReferencia,
		Notas: c.Notas, State: string(e.State),
	}
}

// HandleLabResultCreate crea un resultado de laboratorio para un paciente.
// Delega en LabResultService (Handler -> Service -> Shader/CapabilityGuard -> Core).
// POST /api/v1/patients/:id/lab-results
func HandleLabResultCreate(w http.ResponseWriter, r *http.Request) {
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
	patientID := strings.TrimPrefix(r.URL.Path, "/api/v1/patients/")
	patientID = strings.TrimSuffix(patientID, "/lab-results")

	var req LabResultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "payload invalido")
		return
	}
	if patientID == "" {
		patientID = req.PatientID
	}
	content := shaders.LabResultContent{
		Estudio:      strings.TrimSpace(req.Estudio),
		FechaEstudio: strings.TrimSpace(req.FechaEstudio),
		Resultado:    req.Resultado, Laboratorio: req.Laboratorio,
		Unidades: req.Unidades, ValorReferencia: req.ValorReferencia,
		Notas: req.Notas,
	}

	e, err := deps.LabResultService.Create(tenantID, actorID, patientID, content)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "LAB_RESULT_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": toLabResultItem(e, patientID), "error": nil})
}

// HandleLabResultListByPatient retorna los resultados de laboratorio de un paciente.
// Lee de lab_result_projections via LabResultService (ADR-0022).
// GET /api/v1/patients/:id/lab-results
func HandleLabResultListByPatient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := TenantIDFromContext(r)
	if tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}
	patientID := strings.TrimPrefix(r.URL.Path, "/api/v1/patients/")
	patientID = strings.TrimSuffix(patientID, "/lab-results")
	if patientID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "patient_id requerido")
		return
	}

	projs, err := deps.LabResultService.ListByPatient(tenantID, patientID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al obtener resultados")
		return
	}
	items := make([]LabResultItem, 0, len(projs))
	for _, p := range projs {
		items = append(items, LabResultItem{
			ID: p.EvidenceID, TenantID: p.TenantID, PatientID: p.PatientID,
			Estudio: p.Estudio, FechaEstudio: p.FechaEstudio,
			Resultado: p.Resultado, Laboratorio: p.Laboratorio,
			Unidades: p.Unidades, ValorReferencia: p.ValorReferencia,
			Notas: p.Notas, State: p.State,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"items": items}, "error": nil})
}

// HandleLabResultVoid anula un resultado de laboratorio. Corrección via void (ADR-0006).
// POST /api/v1/lab-results/:id/void
func HandleLabResultVoid(w http.ResponseWriter, r *http.Request) {
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
	labResultID := strings.TrimPrefix(r.URL.Path, "/api/v1/lab-results/")
	labResultID = strings.TrimSuffix(labResultID, "/void")
	if labResultID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "lab_result id requerido")
		return
	}

	voided, err := deps.LabResultService.Void(tenantID, actorID, labResultID)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "LAB_RESULT_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": toItem(voided), "error": nil})
}
