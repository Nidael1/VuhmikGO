package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Nidael1/VuhmikGO/internal/application/ports"
	"github.com/Nidael1/VuhmikGO/internal/shaders"
)

// ConsultationItem es el DTO de respuesta para una consulta.
type ConsultationItem struct {
	ID          string     `json:"id"`
	TenantID    string     `json:"tenant_id"`
	PatientID   string     `json:"patient_id"`
	TA          string     `json:"ta,omitempty"`
	FC          string     `json:"fc,omitempty"`
	FR          string     `json:"fr,omitempty"`
	Temp        string     `json:"temp,omitempty"`
	Peso        string     `json:"peso,omitempty"`
	Talla       string     `json:"talla,omitempty"`
	SAO2        string     `json:"sao2,omitempty"`
	State       string     `json:"state"`
	CreatedAt   time.Time  `json:"created_at"`
	IssuedAt    *time.Time `json:"issued_at,omitempty"`
	TieneReceta bool       `json:"tiene_receta"`
}

func toConsultationItem(p ports.ConsultationProjection) ConsultationItem {
	return ConsultationItem{
		ID:          p.EvidenceID,
		TenantID:    p.TenantID,
		PatientID:   p.PatientID,
		TA:          p.TA,
		FC:          p.FC,
		FR:          p.FR,
		Temp:        p.Temp,
		Peso:        p.Peso,
		Talla:       p.Talla,
		SAO2:        p.SAO2,
		State:       p.State,
		CreatedAt:   p.CreatedAt,
		IssuedAt:    p.IssuedAt,
		TieneReceta: p.TieneReceta,
	}
}

// ConsultationRequest es el payload para crear una consulta.
type ConsultationRequest struct {
	TA          string `json:"ta,omitempty"`
	FC          string `json:"fc,omitempty"`
	FR          string `json:"fr,omitempty"`
	Temp        string `json:"temp,omitempty"`
	Peso        string `json:"peso,omitempty"`
	Talla       string `json:"talla,omitempty"`
	SAO2        string `json:"sao2,omitempty"`
	TieneReceta bool   `json:"tiene_receta"`
}

// HandleConsultationCreate crea y emite una consulta para un paciente.
//
// POST /api/v1/patients/:id/consultations
func HandleConsultationCreate(w http.ResponseWriter, r *http.Request) {
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

	patientID := extractPatientID(r.URL.Path, "/consultations")
	if patientID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "patient_id requerido")
		return
	}

	var req ConsultationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "payload invalido")
		return
	}

	content := shaders.ConsultationContent{
		TA:    req.TA,
		FC:    req.FC,
		FR:    req.FR,
		Temp:  req.Temp,
		Peso:  req.Peso,
		Talla: req.Talla,
		SAO2:  req.SAO2,
	}

	e, err := deps.ConsultationService.Create(tenantID, actorID, patientID, content, req.TieneReceta)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "CONSULTATION_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"data":  map[string]any{"id": e.ID, "state": string(e.State)},
		"error": nil,
	})
}

// HandleConsultationListByPatient lista consultas emitidas de un paciente.
//
// GET /api/v1/patients/:id/consultations
func HandleConsultationListByPatient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := TenantIDFromContext(r)
	if tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}

	patientID := extractPatientID(r.URL.Path, "/consultations")
	items, err := deps.ConsultationService.ListByPatient(tenantID, patientID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al listar consultas")
		return
	}

	dtos := make([]ConsultationItem, 0, len(items))
	for _, p := range items {
		dtos = append(dtos, toConsultationItem(p))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data":  map[string]any{"items": dtos},
		"error": nil,
	})
}

// HandleConsultationListAll lista todas las consultas del tenant.
//
// GET /api/v1/consultations
func HandleConsultationListAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := TenantIDFromContext(r)
	if tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}

	items, err := deps.ConsultationService.ListAll(tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al listar consultas")
		return
	}

	dtos := make([]ConsultationItem, 0, len(items))
	for _, p := range items {
		dtos = append(dtos, toConsultationItem(p))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data":  map[string]any{"items": dtos},
		"error": nil,
	})
}

// HandleConsultationDetail retorna el detalle de una consulta.
//
// GET /api/v1/consultations/:id
func HandleConsultationDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := TenantIDFromContext(r)
	if tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/consultations/")
	p, err := deps.ConsultationService.FindByID(tenantID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "consulta no encontrada")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":  toConsultationItem(p),
		"error": nil,
	})
}
