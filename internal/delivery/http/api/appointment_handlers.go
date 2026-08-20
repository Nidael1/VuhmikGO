package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Nidael1/VuhmikGO/internal/application/ports"
)

// AppointmentItem es el DTO de respuesta para una cita.
type AppointmentItem struct {
	ID             string    `json:"id"`
	PatientID      string    `json:"patient_id"`
	PatientNombre  string    `json:"patient_nombre,omitempty"`
	ScheduledAt    time.Time `json:"scheduled_at"`
	DurationMin    int       `json:"duration_min"`
	Reason         string    `json:"reason"`
	State          string    `json:"state"`
	ConsultationID string    `json:"consultation_id,omitempty"`
	Notes          string    `json:"notes,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

func toAppointmentItem(a ports.Appointment) AppointmentItem {
	return AppointmentItem{
		ID:             a.ID,
		PatientID:      a.PatientID,
		ScheduledAt:    a.ScheduledAt,
		DurationMin:    a.DurationMin,
		Reason:         a.Reason,
		State:          a.State,
		ConsultationID: a.ConsultationID,
		Notes:          a.Notes,
		CreatedAt:      a.CreatedAt,
	}
}

// AppointmentCreateRequest es el payload para crear una cita.
type AppointmentCreateRequest struct {
	PatientID   string `json:"patient_id"`
	ScheduledAt string `json:"scheduled_at"` // RFC3339
	DurationMin int    `json:"duration_min"`
	Reason      string `json:"reason"`
	Notes       string `json:"notes"`
}

// HandleAppointmentCreate crea una cita agendada.
//
// POST /api/v1/appointments
func HandleAppointmentCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := TenantIDFromContext(r)
	if tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}

	var req AppointmentCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "payload invalido")
		return
	}

	if req.PatientID == "" || req.ScheduledAt == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "patient_id y scheduled_at son obligatorios")
		return
	}

	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_DATE", "scheduled_at debe ser RFC3339")
		return
	}
	if scheduledAt.Before(time.Now()) {
		writeError(w, http.StatusBadRequest, "INVALID_DATE", "scheduled_at debe ser en el futuro")
		return
	}

	duration := req.DurationMin
	if duration <= 0 {
		duration = 30
	}
	if duration < 5 || duration > 480 {
		writeError(w, http.StatusBadRequest, "INVALID_DURATION", "duration_min debe estar entre 5 y 480")
		return
	}

	// Verificar que el paciente pertenece al tenant.
	if _, err := deps.PatientRepo.FindByID(tenantID, req.PatientID); err != nil {
		writeError(w, http.StatusBadRequest, "PATIENT_NOT_FOUND", "paciente no encontrado")
		return
	}

	now := time.Now()
	a := ports.Appointment{
		ID:          fmt.Sprintf("apt-%s-%s", tenantID[:4], now.Format("20060102150405.000")),
		TenantID:    tenantID,
		PatientID:   req.PatientID,
		ScheduledAt: scheduledAt,
		DurationMin: duration,
		Reason:      req.Reason,
		State:       "scheduled",
		Notes:       req.Notes,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := deps.AppointmentRepo.Create(a); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al crear la cita")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"data":  map[string]any{"id": a.ID, "state": a.State},
		"error": nil,
	})
}

// HandleAppointmentListAll lista todas las citas del tenant.
//
// GET /api/v1/appointments
func HandleAppointmentListAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := TenantIDFromContext(r)
	if tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}

	items, err := deps.AppointmentRepo.ListByTenant(tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al listar citas")
		return
	}

	dtos := make([]AppointmentItem, 0, len(items))
	for _, a := range items {
		dto := toAppointmentItem(a)
		if pat, err := deps.PatientRepo.FindByID(tenantID, a.PatientID); err == nil {
			dto.PatientNombre = pat.Nombre
		}
		dtos = append(dtos, dto)
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"items": dtos}, "error": nil})
}

// HandleAppointmentToday lista las citas del día.
//
// GET /api/v1/appointments/today?date=YYYY-MM-DD
func HandleAppointmentToday(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := TenantIDFromContext(r)
	if tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}

	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	items, err := deps.AppointmentRepo.ListToday(tenantID, date)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al listar citas del dia")
		return
	}

	dtos := make([]AppointmentItem, 0, len(items))
	for _, a := range items {
		dto := toAppointmentItem(a)
		if pat, err := deps.PatientRepo.FindByID(tenantID, a.PatientID); err == nil {
			dto.PatientNombre = pat.Nombre
		}
		dtos = append(dtos, dto)
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"items": dtos}, "error": nil})
}

// HandleAppointmentListByPatient lista las citas de un paciente.
//
// GET /api/v1/patients/:id/appointments
func HandleAppointmentListByPatient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := TenantIDFromContext(r)
	if tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}

	patientID := extractPatientID(r.URL.Path, "/appointments")
	items, err := deps.AppointmentRepo.ListByPatient(tenantID, patientID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al listar citas del paciente")
		return
	}

	dtos := make([]AppointmentItem, 0, len(items))
	for _, a := range items {
		dtos = append(dtos, toAppointmentItem(a))
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"items": dtos}, "error": nil})
}

// HandleAppointmentUpdateState cambia el estado de una cita (cancelled, no_show).
//
// PATCH /api/v1/appointments/:id/state
func HandleAppointmentUpdateState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := TenantIDFromContext(r)
	if tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}

	id := extractAppointmentID(r.URL.Path, "/state")

	var body struct {
		State string `json:"state"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "payload invalido")
		return
	}

	allowed := map[string]bool{"cancelled": true, "no_show": true, "completed": true}
	if !allowed[body.State] {
		writeError(w, http.StatusBadRequest, "INVALID_STATE", "state debe ser cancelled, no_show o completed")
		return
	}

	a, err := deps.AppointmentRepo.FindByID(tenantID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "cita no encontrada")
		return
	}
	if a.State != "scheduled" {
		writeError(w, http.StatusUnprocessableEntity, "INVALID_STATE", "solo se pueden cancelar citas en estado scheduled")
		return
	}

	if err := deps.AppointmentRepo.UpdateState(tenantID, id, body.State); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al actualizar la cita")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"id": id, "state": body.State}, "error": nil})
}

// extractAppointmentID extrae el ID de la cita del path, removiendo el sufijo dado.
func extractAppointmentID(path, suffix string) string {
	path = strings.TrimPrefix(path, "/api/v1/appointments/")
	path = strings.TrimSuffix(path, suffix)
	return strings.Trim(path, "/")
}
