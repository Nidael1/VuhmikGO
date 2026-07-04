package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Nidael1/VuhmikGO/internal/application/ports"
	"github.com/Nidael1/VuhmikGO/internal/auth"
	"github.com/Nidael1/VuhmikGO/internal/shaders"
)

// PrescriptionItem es el DTO de respuesta para una receta.
type PrescriptionItem struct {
	ID                  string     `json:"id"`
	TenantID            string     `json:"tenant_id"`
	PatientID           string     `json:"patient_id"`
	MedicamentoGenerico string     `json:"medicamento_generico"`
	Dosis               string     `json:"dosis"`
	Diagnostico         string     `json:"diagnostico,omitempty"`
	Indicaciones        string     `json:"indicaciones,omitempty"`
	Seguimiento         string     `json:"seguimiento,omitempty"`
	State               string     `json:"state"`
	CreatedAt           time.Time  `json:"created_at"`
	IssuedAt            *time.Time `json:"issued_at,omitempty"`
	ConsultationID      string     `json:"consultation_id,omitempty"`
}

func toPrescriptionItem(p ports.PrescriptionProjection) PrescriptionItem {
	return PrescriptionItem{
		ID:                  p.EvidenceID,
		TenantID:            p.TenantID,
		PatientID:           p.PatientID,
		MedicamentoGenerico: p.MedicamentoGenerico,
		Dosis:               p.Dosis,
		Diagnostico:         p.Diagnostico,
		Indicaciones:        p.Indicaciones,
		Seguimiento:         p.Seguimiento,
		State:               p.State,
		CreatedAt:           p.CreatedAt,
		ConsultationID:      p.ConsultationID,
		IssuedAt:            p.IssuedAt,
	}
}

// PrescriptionRequest es el payload para crear una receta draft.
type PrescriptionRequest struct {
	MedicamentoGenerico string `json:"medicamento_generico"`
	Dosis               string `json:"dosis"`
	Diagnostico         string `json:"diagnostico,omitempty"`
	Indicaciones        string `json:"indicaciones,omitempty"`
	Seguimiento         string `json:"seguimiento,omitempty"`
	ConsultationID      string `json:"consultation_id,omitempty"`
}

// HandlePrescriptionCreate crea un borrador de receta para un paciente.
//
// POST /api/v1/patients/:id/prescriptions
func HandlePrescriptionCreate(w http.ResponseWriter, r *http.Request) {
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

	patientID := extractPatientID(r.URL.Path, "/prescriptions")
	if patientID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "patient_id requerido")
		return
	}

	var req PrescriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "payload invalido")
		return
	}
	if strings.TrimSpace(req.MedicamentoGenerico) == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "medicamento_generico es obligatorio")
		return
	}
	if strings.TrimSpace(req.Dosis) == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "dosis es obligatoria")
		return
	}

	content := shaders.PrescriptionContent{
		MedicamentoGenerico: req.MedicamentoGenerico,
		Dosis:               req.Dosis,
		Diagnostico:         req.Diagnostico,
		Indicaciones:        req.Indicaciones,
		Seguimiento:         req.Seguimiento,
	}

	e, err := deps.PrescriptionService.CreateDraft(tenantID, actorID, patientID, req.ConsultationID, content)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "PRESCRIPTION_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"data":  map[string]any{"id": e.ID, "state": string(e.State)},
		"error": nil,
	})
}

// HandlePrescriptionEmit emite una receta draft — adquiere validez legal.
// Valida campos minimos NOM-024: cedula + especialidad del perfil del medico.
//
// POST /api/v1/prescriptions/:id/emit
func HandlePrescriptionEmit(w http.ResponseWriter, r *http.Request) {
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

	prescriptionID := strings.TrimPrefix(r.URL.Path, "/api/v1/prescriptions/")
	prescriptionID = strings.TrimSuffix(prescriptionID, "/emit")
	if prescriptionID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "prescription_id requerido")
		return
	}

	// Obtener perfil del medico para cedula y especialidad (NOM-024)
	profile, err := deps.ProfileRepo.Get(actorID)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "PROFILE_INCOMPLETE",
			"completa tu perfil profesional antes de emitir una receta")
		return
	}
	// Validación NOM-024 delegada al shader mx_medical (ADR-0002, issue #202).
	// El handler es transporte puro; la regla de dominio vive en el Shader.
	if err := shaders.ValidateMxMedicalProfile(shaders.MxMedicalProfile{
		CedulaProfesional: profile.CedulaProfesional,
		Especialidad:      profile.Especialidad,
	}); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "PROFILE_INCOMPLETE", err.Error())
		return
	}

	e, err := deps.PrescriptionService.Emit(tenantID, actorID, prescriptionID, profile)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "PRESCRIPTION_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":  map[string]any{"id": e.ID, "state": string(e.State), "issued_at": e.IssuedAt},
		"error": nil,
	})
}

// HandlePrescriptionVoid anula una receta emitida.
//
// POST /api/v1/prescriptions/:id/void
func HandlePrescriptionVoid(w http.ResponseWriter, r *http.Request) {
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

	prescriptionID := strings.TrimPrefix(r.URL.Path, "/api/v1/prescriptions/")
	prescriptionID = strings.TrimSuffix(prescriptionID, "/void")

	voided, err := deps.PrescriptionService.Void(tenantID, actorID, prescriptionID)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "PRESCRIPTION_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":  map[string]any{"id": voided.ID, "state": string(voided.State)},
		"error": nil,
	})
}

// HandlePrescriptionListByPatient lista recetas emitidas de un paciente.
//
// GET /api/v1/patients/:id/prescriptions
func HandlePrescriptionListByPatient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := TenantIDFromContext(r)
	if tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}

	patientID := extractPatientID(r.URL.Path, "/prescriptions")
	items, err := deps.PrescriptionService.ListByPatient(tenantID, patientID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al listar recetas")
		return
	}

	dtos := make([]PrescriptionItem, 0, len(items))
	for _, p := range items {
		dtos = append(dtos, toPrescriptionItem(p))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data":  map[string]any{"items": dtos},
		"error": nil,
	})
}

// HandlePrescriptionListAll lista todas las recetas del tenant.
// Para la vista global de recetas en el sidebar.
//
// GET /api/v1/prescriptions
func HandlePrescriptionListAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := TenantIDFromContext(r)
	if tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}

	items, err := deps.PrescriptionService.ListAll(tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al listar recetas")
		return
	}

	dtos := make([]PrescriptionItem, 0, len(items))
	for _, p := range items {
		dtos = append(dtos, toPrescriptionItem(p))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data":  map[string]any{"items": dtos},
		"error": nil,
	})
}

// HandlePrescriptionDetail retorna el detalle de una receta.
//
// GET /api/v1/prescriptions/:id
func HandlePrescriptionDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := TenantIDFromContext(r)
	if tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/prescriptions/")
	p, err := deps.PrescriptionService.FindByID(tenantID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "receta no encontrada")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":  toPrescriptionItem(p),
		"error": nil,
	})
}

// claimsKey ya definida en context.go — no redefinir aquí
var _ = auth.Claims{}
