package api

import (
	"encoding/json"
	"net/http"
)

// --- DTOs admin ---

type adminModuleItem struct {
	ModuleID    string  `json:"ModuleID"`
	Descripcion string  `json:"Descripcion"`
	Active      bool    `json:"Active"`
	Plan        string  `json:"Plan"`
	Costo       float64 `json:"Costo"`
}

type adminTenantItem struct {
	TenantID    string            `json:"tenant_id"`
	Email       string            `json:"email"`
	IsAdmin     bool              `json:"is_admin"`
	IsSuspended bool              `json:"is_suspended"`
	Modules     []adminModuleItem `json:"modules"`
}

// HandleAdminTenants lista todos los tenants con sus módulos.
//
// GET /api/v1/admin/tenants
func HandleAdminTenants(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	users, err := deps.UserRepo.FindAll()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al listar usuarios")
		return
	}
	items := make([]adminTenantItem, 0, len(users))
	for _, u := range users {
		// El admin no es un médico — no aparece en el panel de control
		if u.IsAdmin {
			continue
		}
		mods, _ := deps.CapabilityRepo.ListByTenant(u.TenantID, "medico")
		modItems := make([]adminModuleItem, 0, len(mods))
		for _, m := range mods {
			modItems = append(modItems, adminModuleItem{
				ModuleID:    m.ModuleID,
				Descripcion: m.Descripcion,
				Active:      m.Active,
				Plan:        m.Plan,
				Costo:       m.Costo,
			})
		}
		items = append(items, adminTenantItem{
			TenantID:    u.TenantID,
			Email:       u.Email,
			IsAdmin:     u.IsAdmin,
			IsSuspended: u.IsSuspended,
			Modules:     modItems,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data":  map[string]any{"items": items},
		"error": nil,
	})
}

// CapabilityToggleRequest es el payload para activar/desactivar un módulo.
type CapabilityToggleRequest struct {
	TenantID string `json:"tenant_id"`
	ModuleID string `json:"module_id"`
	Active   bool   `json:"active"`
}

// HandleAdminCapabilityToggle activa o desactiva un módulo para un tenant.
//
// POST /api/v1/admin/capabilities
func HandleAdminCapabilityToggle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	var req CapabilityToggleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "payload invalido")
		return
	}
	if req.TenantID == "" || req.ModuleID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "tenant_id y module_id son obligatorios")
		return
	}
	if req.Active {
		if err := deps.CapabilityRepo.Activate(req.TenantID, req.ModuleID, "basico", 0); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
			return
		}
	} else {
		if err := deps.CapabilityRepo.Deactivate(req.TenantID, req.ModuleID); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]bool{"ok": true}, "error": nil})
}

// HandleAdminSuspend suspende o reactiva una cuenta de médico.
//
// POST /api/v1/admin/suspend
func HandleAdminSuspend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	var req struct {
		UserID      string `json:"user_id"`
		IsSuspended bool   `json:"is_suspended"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "payload invalido")
		return
	}
	if err := deps.UserRepo.SetSuspended(req.UserID, req.IsSuspended); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]bool{"ok": true}, "error": nil})
}
