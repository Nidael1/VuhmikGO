package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Nidael1/VuhmikGO/internal/application/ports"
	"github.com/Nidael1/VuhmikGO/internal/auth"
)

// HandleGetProfile retorna el perfil profesional del actor autenticado.
//
// GET /api/v1/profile
func HandleGetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	claims, ok := r.Context().Value(claimsKey{}).(*auth.Claims)
	if !ok || claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}
	profile, err := deps.ProfileRepo.Get(claims.ActorID)
	if err != nil {
		// Si no existe aun, devolver perfil vacio
		profile = ports.Profile{
			UserID:   claims.ActorID,
			TenantID: claims.TenantID,
			Rubro:    "medicine",
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": profile, "error": nil})
}

// ProfileUpdateRequest es el payload para actualizar el perfil.
type ProfileUpdateRequest struct {
	NombreCompleto    string `json:"nombre_completo"`
	CedulaProfesional string `json:"cedula_profesional"`
	Especialidad      string `json:"especialidad"`
	Universidad       string `json:"universidad"`
	Direccion         string `json:"direccion"`
	Telefono          string `json:"telefono"`
}

// HandleUpdateProfile actualiza el perfil profesional del actor autenticado.
//
// PUT /api/v1/profile
func HandleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	claims, ok := r.Context().Value(claimsKey{}).(*auth.Claims)
	if !ok || claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}
	var req ProfileUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "payload invalido")
		return
	}
	p := ports.Profile{
		UserID:            claims.ActorID,
		TenantID:          claims.TenantID,
		Rubro:             "medicine",
		NombreCompleto:    strings.TrimSpace(req.NombreCompleto),
		CedulaProfesional: strings.TrimSpace(req.CedulaProfesional),
		Especialidad:      strings.TrimSpace(req.Especialidad),
		Universidad:       strings.TrimSpace(req.Universidad),
		Direccion:         strings.TrimSpace(req.Direccion),
		Telefono:          strings.TrimSpace(req.Telefono),
	}
	if err := deps.ProfileRepo.Upsert(p); err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", "error al guardar perfil")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": p, "error": nil})
}
