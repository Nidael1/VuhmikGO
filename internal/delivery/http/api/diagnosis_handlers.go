package api

import (
	"encoding/json"
	"net/http"
	"strings"

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
// Delega en DiagnosisService (Handler -> Service -> Shader/CapabilityGuard -> Core).
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
		Tipo:        req.Tipo, EstadoProblema: req.EstadoProblema,
		FechaInicio: req.FechaInicio, Notas: req.Notas,
	}

	e, err := deps.DiagnosisService.Create(tenantID, actorID, patientID, content)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "DIAGNOSIS_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": toDiagnosisItem(e, patientID), "error": nil})
}

// HandleDiagnosisListByPatient retorna los diagnósticos activos de un paciente.
// Lee de diagnosis_projections via DiagnosisService (ADR-0022).
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

	projs, err := deps.DiagnosisService.ListByPatient(tenantID, patientID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al obtener diagnosticos")
		return
	}
	items := make([]DiagnosisItem, 0, len(projs))
	for _, p := range projs {
		items = append(items, DiagnosisItem{
			ID: p.EvidenceID, TenantID: p.TenantID, PatientID: p.PatientID,
			Descripcion: p.Descripcion, CodigoCIE10: p.CodigoCIE10,
			Tipo: p.Tipo, EstadoProblema: p.EstadoProblema,
			FechaInicio: p.FechaInicio, Notas: p.Notas,
			State: p.State,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"items": items}, "error": nil})
}

// HandleDiagnosisVoid anula un diagnóstico. Corrección via void (ADR-0006).
// POST /api/v1/diagnoses/:id/void
func HandleDiagnosisVoid(w http.ResponseWriter, r *http.Request) {
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
	diagnosisID := strings.TrimPrefix(r.URL.Path, "/api/v1/diagnoses/")
	diagnosisID = strings.TrimSuffix(diagnosisID, "/void")
	if diagnosisID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "diagnosis id requerido")
		return
	}

	voided, err := deps.DiagnosisService.Void(tenantID, actorID, diagnosisID)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "DIAGNOSIS_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": toItem(voided), "error": nil})
}
