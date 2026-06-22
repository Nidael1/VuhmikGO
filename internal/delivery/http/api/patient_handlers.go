package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Patient representa un paciente del consultorio.
// Pertenece a Asteroides (CRM) — no al Core.
// Campos minimos segun NOM-004-SSA3-2012, numeral 5.9.
type Patient struct {
	ID              string    `json:"id"`
	TenantID        string    `json:"tenant_id"`
	Nombre          string    `json:"nombre"`
	FechaNacimiento string    `json:"fecha_nacimiento"` // YYYY-MM-DD
	Sexo            string    `json:"sexo"`             // M, F, I
	NumExpediente   string    `json:"num_expediente"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// patientStore es el store en memoria de pacientes para demo.
var (
	patientStore   = map[string]*Patient{}
	patientStoreMu sync.RWMutex
	patientCounter = 0
)

func nextExpediente(tenantID string) string {
	patientCounter++
	return fmt.Sprintf("EXP-%04d", patientCounter)
}

// PatientRequest es el payload para crear o actualizar un paciente.
type PatientRequest struct {
	Nombre          string `json:"nombre"`
	FechaNacimiento string `json:"fecha_nacimiento"`
	Sexo            string `json:"sexo"`
}

// HandlePatientList retorna todos los pacientes del tenant.
//
// GET /api/v1/patients
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
	patientStoreMu.RLock()
	defer patientStoreMu.RUnlock()
	var items []*Patient
	for _, p := range patientStore {
		if p.TenantID == tenantID {
			items = append(items, p)
		}
	}
	if items == nil {
		items = []*Patient{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"items": items}, "error": nil})
}

// HandlePatientCreate crea un nuevo paciente.
//
// POST /api/v1/patients
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
	now := time.Now().UTC()
	patientStoreMu.Lock()
	expediente := nextExpediente(tenantID)
	p := &Patient{
		ID:              "pac-" + tenantID + "-" + now.Format("20060102150405"),
		TenantID:        tenantID,
		Nombre:          req.Nombre,
		FechaNacimiento: req.FechaNacimiento,
		Sexo:            req.Sexo,
		NumExpediente:   expediente,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	patientStore[p.ID] = p
	patientStoreMu.Unlock()
	writeJSON(w, http.StatusCreated, map[string]any{"data": p, "error": nil})
}

// HandlePatientDetail retorna el detalle de un paciente.
//
// GET /api/v1/patients/:id
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
	patientStoreMu.RLock()
	p, ok := patientStore[id]
	patientStoreMu.RUnlock()
	if !ok || p.TenantID != tenantID {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "paciente no encontrado")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": p, "error": nil})
}

// HandlePatientUpdate actualiza los datos de un paciente.
//
// PUT /api/v1/patients/:id
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
	patientStoreMu.Lock()
	defer patientStoreMu.Unlock()
	p, ok := patientStore[id]
	if !ok || p.TenantID != tenantID {
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
	writeJSON(w, http.StatusOK, map[string]any{"data": p, "error": nil})
}
