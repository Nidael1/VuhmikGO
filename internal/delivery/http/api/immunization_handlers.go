package api

import (
	"encoding/json"
	"net/http"
	"strings"

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
// Delega en ImmunizationService (Handler -> Service -> Shader/CapabilityGuard -> Core).
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
		Vacuna:          strings.TrimSpace(req.Vacuna),
		FechaAplicacion: strings.TrimSpace(req.FechaAplicacion),
		Lote:            req.Lote, Dosis: req.Dosis, Via: req.Via,
		AplicadaPor: req.AplicadaPor, Notas: req.Notas,
	}

	e, err := deps.ImmunizationService.Create(tenantID, actorID, patientID, content)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "IMMUNIZATION_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": toImmunizationItem(e, patientID), "error": nil})
}

// HandleImmunizationListByPatient retorna las vacunas de un paciente.
// Lee de immunization_projections via ImmunizationService (ADR-0022).
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

	projs, err := deps.ImmunizationService.ListByPatient(tenantID, patientID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al obtener vacunas")
		return
	}
	items := make([]ImmunizationItem, 0, len(projs))
	for _, p := range projs {
		items = append(items, ImmunizationItem{
			ID: p.EvidenceID, TenantID: p.TenantID, PatientID: p.PatientID,
			Vacuna: p.Vacuna, FechaAplicacion: p.FechaAplicacion,
			Lote: p.Lote, Dosis: p.Dosis, Via: p.Via,
			AplicadaPor: p.AplicadaPor, Notas: p.Notas,
			State: p.State,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"items": items}, "error": nil})
}

// HandleImmunizationVoid anula una inmunización. Corrección via void (ADR-0006).
// POST /api/v1/immunizations/:id/void
func HandleImmunizationVoid(w http.ResponseWriter, r *http.Request) {
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
	immunizationID := strings.TrimPrefix(r.URL.Path, "/api/v1/immunizations/")
	immunizationID = strings.TrimSuffix(immunizationID, "/void")
	if immunizationID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "immunization id requerido")
		return
	}

	voided, err := deps.ImmunizationService.Void(tenantID, actorID, immunizationID)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "IMMUNIZATION_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": toItem(voided), "error": nil})
}
