package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

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
		Estudio: strings.TrimSpace(req.Estudio),
		FechaEstudio: strings.TrimSpace(req.FechaEstudio),
		Resultado: req.Resultado, Laboratorio: req.Laboratorio,
		Unidades: req.Unidades, ValorReferencia: req.ValorReferencia,
		Notas: req.Notas,
	}
	if err := shaders.ValidateLabResultContent(content); err != nil {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", err.Error())
		return
	}
	blob, err := shaders.BuildLabResultBlob(content)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al construir blob")
		return
	}
	now := time.Now().UTC()
	id := "lab-" + now.Format("20060102150405.000")
	e := evidence.Evidence{
		ID: id, TenantID: tenantID, SubjectRef: patientID,
		Content: blob, State: evidence.StateDraft, CreatedAt: now,
	}
	if err := deps.EvidenceRepo.Create(e); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al crear resultado")
		return
	}
	e.State = evidence.StateIssued
	e.IssuedAt = &now
	if err := deps.EvidenceRepo.Update(tenantID, e); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al emitir resultado")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": toLabResultItem(e, patientID), "error": nil})
}

// HandleLabResultListByPatient retorna los resultados de laboratorio de un paciente.
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
	allEvidence, err := deps.EvidenceRepo.FindAll(tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al obtener resultados")
		return
	}
	items := make([]LabResultItem, 0)
	for _, e := range allEvidence {
		if e.SubjectRef != patientID {
			continue
		}
		var c shaders.LabResultContent
		if err := shaders.ParseLabResultBlob(e.Content, &c); err != nil {
			continue
		}
		if c.Type != "lab_result" || e.State == evidence.StateVoided {
			continue
		}
		items = append(items, toLabResultItem(e, patientID))
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"items": items}, "error": nil})
}
