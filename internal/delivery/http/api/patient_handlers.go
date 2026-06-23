package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Nidael1/VuhmikGO/internal/infrastructure/postgres"
)

// PatientItem es el DTO de respuesta para paciente.
type PatientItem struct {
	ID              string    `json:"id"`
	TenantID        string    `json:"tenant_id"`
	Nombre          string    `json:"nombre"`
	FechaNacimiento string    `json:"fecha_nacimiento"`
	Sexo            string    `json:"sexo"`
	NumExpediente   string    `json:"num_expediente"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func toPatientItem(p postgres.Patient) PatientItem {
	return PatientItem{
		ID:              p.ID,
		TenantID:        p.TenantID,
		Nombre:          p.Nombre,
		FechaNacimiento: p.FechaNacimiento,
		Sexo:            p.Sexo,
		NumExpediente:   p.NumExpediente,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}
}

// PatientRequest es el payload para crear o actualizar un paciente.
type PatientRequest struct {
	Nombre          string `json:"nombre"`
	FechaNacimiento string `json:"fecha_nacimiento"`
	Sexo            string `json:"sexo"`
}

// HandlePatientList retorna todos los pacientes del tenant.
func HandlePatientList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := TenantIDFromContext(r)
	if tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}
	patients, err := deps.PatientRepo.FindAll(tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al listar pacientes")
		return
	}
	items := make([]PatientItem, 0, len(patients))
	for _, p := range patients {
		items = append(items, toPatientItem(p))
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"items": items}, "error": nil})
}

// HandlePatientCreate crea un nuevo paciente.
func HandlePatientCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := TenantIDFromContext(r)
	if tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}
	var req PatientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "payload invalido")
		return
	}
	if strings.TrimSpace(req.Nombre) == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "nombre es obligatorio")
		return
	}
	if strings.TrimSpace(req.FechaNacimiento) == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "fecha_nacimiento es obligatoria")
		return
	}
	if req.Sexo != "M" && req.Sexo != "F" && req.Sexo != "I" {
		writeError(w, http.StatusBadRequest, "INVALID_FIELDS", "sexo debe ser M, F o I")
		return
	}
	expediente, err := deps.PatientRepo.NextExpediente(tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al generar expediente")
		return
	}
	now := time.Now().UTC()
	p := postgres.Patient{
		ID:              "pac-" + tenantID + "-" + now.Format("20060102150405"),
		TenantID:        tenantID,
		Nombre:          strings.ToUpper(strings.TrimSpace(req.Nombre)),
		FechaNacimiento: req.FechaNacimiento,
		Sexo:            req.Sexo,
		NumExpediente:   expediente,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := deps.PatientRepo.Create(p); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al crear paciente")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": toPatientItem(p), "error": nil})
}

// HandlePatientDetail retorna el detalle de un paciente.
func HandlePatientDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := TenantIDFromContext(r)
	if tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/patients/")
	p, err := deps.PatientRepo.FindByID(tenantID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "paciente no encontrado")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": toPatientItem(p), "error": nil})
}

// HandlePatientUpdate actualiza los datos de un paciente.
func HandlePatientUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := TenantIDFromContext(r)
	if tenantID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/patients/")
	id = strings.TrimSuffix(id, "/edit")
	var req PatientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "payload invalido")
		return
	}
	p, err := deps.PatientRepo.FindByID(tenantID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "paciente no encontrado")
		return
	}
	if strings.TrimSpace(req.Nombre) != "" {
		p.Nombre = req.Nombre
	}
	if strings.TrimSpace(req.FechaNacimiento) != "" {
		p.FechaNacimiento = req.FechaNacimiento
	}
	if req.Sexo == "M" || req.Sexo == "F" || req.Sexo == "I" {
		p.Sexo = req.Sexo
	}
	p.UpdatedAt = time.Now().UTC()
	if err := deps.PatientRepo.Update(tenantID, p); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al actualizar paciente")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": toPatientItem(p), "error": nil})
}
