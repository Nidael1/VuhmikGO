package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Nidael1/VuhmikGO/internal/core/evidence"
	"github.com/Nidael1/VuhmikGO/internal/shaders"
)

type ImmunizationRequest struct {
	PatientID       string `json:"patient_id"`
	Vacuna          string `json:"vacuna"`
	FechaAplicacion string `json:"fecha_aplicacion"`
	Lote            string `json:"lote,omitempty"`
	Dosis           string `json:"dosis,omitempty"`
	Via             string `json:"via,omitempty"`
	AplicadaPor     string `json:"aplicada_por,omitempty"`
	Notas           string `json:"notas,omitempty"`
}

type ImmunizationItem struct {
	ID              string `json:"id"`
	TenantID        string `json:"tenant_id"`
	PatientID       string `json:"patient_id"`
	Vacuna          string `json:"vacuna"`
	FechaAplicacion string `json:"fecha_aplicacion"`
	Lote            string `json:"lote,omitempty"`
	Dosis           string `json:"dosis,omitempty"`
	Via             string `json:"via,omitempty"`
	AplicadaPor     string `json:"aplicada_por,omitempty"`
	Notas           string `json:"notas,omitempty"`
	State           string `json:"state"`
}

func toImmunizationItem(e evidence.Evidence, patientID string) ImmunizationItem {
	var c shaders.ImmunizationContent
	_ = shaders.ParseImmunizationBlob(e.Content, &c)
	return ImmunizationItem{
		ID: e.ID, TenantID: e.TenantID, PatientID: patientID,
		Vacuna: c.Vacuna, FechaAplicacion: c.FechaAplicacion,
		Lote: c.Lote, Dosis: c.Dosis, Via: c.Via,
		AplicadaPor: c.AplicadaPor, Notas: c.Notas,
		State: string(e.State),
	}
}

// HandleImmunizationCreate crea un registro de vacuna para un paciente.
// POST /api/v1/patients/:id/immunizations
func HandleImmunizationCreate(w http.ResponseWriter, r *http.Request) {
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
	patientID = strings.TrimSuffix(patientID, "/immunizations")

	var req ImmunizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "payload invalido")
		return
	}
	if patientID == "" {
		patientID = req.PatientID
	}
	content := shaders.ImmunizationContent{
		Vacuna: strings.TrimSpace(req.Vacuna),
		FechaAplicacion: strings.TrimSpace(req.FechaAplicacion),
		Lote: req.Lote, Dosis: req.Dosis, Via: req.Via,
		AplicadaPor: req.AplicadaPor, Notas: req.Notas,
	}
	if err := shaders.ValidateImmunizationContent(content); err != nil {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", err.Error())
		return
	}
	blob, err := shaders.BuildImmunizationBlob(content)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al construir blob")
		return
	}
	now := time.Now().UTC()
	id := "imm-" + now.Format("20060102150405.000")
	e := evidence.Evidence{
		ID: id, TenantID: tenantID, SubjectRef: patientID,
		Content: blob, State: evidence.StateDraft, CreatedAt: now,
	}
	if err := deps.EvidenceRepo.Create(e); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al crear vacuna")
		return
	}
	e.State = evidence.StateIssued
	e.IssuedAt = &now
	if err := deps.EvidenceRepo.Update(tenantID, e); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al emitir vacuna")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": toImmunizationItem(e, patientID), "error": nil})
}

// HandleImmunizationListByPatient retorna las vacunas de un paciente.
// GET /api/v1/patients/:id/immunizations
func HandleImmunizationListByPatient(w http.ResponseWriter, r *http.Request) {
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
	patientID = strings.TrimSuffix(patientID, "/immunizations")
	if patientID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "patient_id requerido")
		return
	}
	allEvidence, err := deps.EvidenceRepo.FindAll(tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al obtener vacunas")
		return
	}
	items := make([]ImmunizationItem, 0)
	for _, e := range allEvidence {
		if e.SubjectRef != patientID {
			continue
		}
		var c shaders.ImmunizationContent
		if err := shaders.ParseImmunizationBlob(e.Content, &c); err != nil {
			continue
		}
		if c.Type != "immunization" || e.State == evidence.StateVoided {
			continue
		}
		items = append(items, toImmunizationItem(e, patientID))
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"items": items}, "error": nil})
}
