package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Nidael1/VuhmikGO/internal/core/evidence"
	"github.com/Nidael1/VuhmikGO/internal/shaders"
)

type DiagnosisRequest struct {
	PatientID      string `json:"patient_id"`
	Descripcion    string `json:"descripcion"`
	CodigoCIE10    string `json:"codigo_cie10,omitempty"`
	Tipo           string `json:"tipo,omitempty"`
	EstadoProblema string `json:"estado_problema,omitempty"`
	FechaInicio    string `json:"fecha_inicio,omitempty"`
	Notas          string `json:"notas,omitempty"`
}

type DiagnosisItem struct {
	ID             string `json:"id"`
	TenantID       string `json:"tenant_id"`
	PatientID      string `json:"patient_id"`
	Descripcion    string `json:"descripcion"`
	CodigoCIE10    string `json:"codigo_cie10,omitempty"`
	Tipo           string `json:"tipo,omitempty"`
	EstadoProblema string `json:"estado_problema,omitempty"`
	FechaInicio    string `json:"fecha_inicio,omitempty"`
	Notas          string `json:"notas,omitempty"`
	State          string `json:"state"`
}

func toDiagnosisItem(e evidence.Evidence, patientID string) DiagnosisItem {
	var c shaders.DiagnosisContent
	_ = shaders.ParseDiagnosisBlob(e.Content, &c)
	return DiagnosisItem{
		ID: e.ID, TenantID: e.TenantID, PatientID: patientID,
		Descripcion: c.Descripcion, CodigoCIE10: c.CodigoCIE10,
		Tipo: c.Tipo, EstadoProblema: c.EstadoProblema,
		FechaInicio: c.FechaInicio, Notas: c.Notas,
		State: string(e.State),
	}
}

// HandleDiagnosisCreate crea un diagnóstico para un paciente.
// POST /api/v1/patients/:id/diagnoses
func HandleDiagnosisCreate(w http.ResponseWriter, r *http.Request) {
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
	patientID = strings.TrimSuffix(patientID, "/diagnoses")

	var req DiagnosisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "payload invalido")
		return
	}
	if patientID == "" {
		patientID = req.PatientID
	}

	content := shaders.DiagnosisContent{
		Descripcion: strings.TrimSpace(req.Descripcion),
		CodigoCIE10: strings.TrimSpace(req.CodigoCIE10),
		Tipo: req.Tipo, EstadoProblema: req.EstadoProblema,
		FechaInicio: req.FechaInicio, Notas: req.Notas,
	}
	if err := shaders.ValidateDiagnosisContent(content); err != nil {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", err.Error())
		return
	}
	blob, err := shaders.BuildDiagnosisBlob(content)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al construir blob")
		return
	}
	now := time.Now().UTC()
	id := "diag-" + now.Format("20060102150405.000")
	e := evidence.Evidence{
		ID: id, TenantID: tenantID, SubjectRef: patientID,
		Content: blob, State: evidence.StateDraft, CreatedAt: now,
	}
	if err := deps.EvidenceRepo.Create(e); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al crear diagnostico")
		return
	}
	e.State = evidence.StateIssued
	e.IssuedAt = &now
	if err := deps.EvidenceRepo.Update(tenantID, e); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al emitir diagnostico")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": toDiagnosisItem(e, patientID), "error": nil})
}

// HandleDiagnosisListByPatient retorna los diagnósticos activos de un paciente.
// GET /api/v1/patients/:id/diagnoses
func HandleDiagnosisListByPatient(w http.ResponseWriter, r *http.Request) {
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
	patientID = strings.TrimSuffix(patientID, "/diagnoses")
	if patientID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "patient_id requerido")
		return
	}
	allEvidence, err := deps.EvidenceRepo.FindAll(tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al obtener diagnosticos")
		return
	}
	items := make([]DiagnosisItem, 0)
	for _, e := range allEvidence {
		if e.SubjectRef != patientID {
			continue
		}
		var c shaders.DiagnosisContent
		if err := shaders.ParseDiagnosisBlob(e.Content, &c); err != nil {
			continue
		}
		if c.Type != "diagnosis" || e.State == evidence.StateVoided {
			continue
		}
		items = append(items, toDiagnosisItem(e, patientID))
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"items": items}, "error": nil})
}
