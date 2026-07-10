package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Nidael1/VuhmikGO/internal/application/ports"
	"github.com/Nidael1/VuhmikGO/internal/auth"
	"github.com/Nidael1/VuhmikGO/internal/infrastructure/postgres"
	"github.com/Nidael1/VuhmikGO/internal/observability"
	"github.com/Nidael1/VuhmikGO/internal/shaders"
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
	BillingMode string            `json:"billing_mode"`
	MonthlyFee  float64           `json:"monthly_fee"`
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
			BillingMode: u.BillingMode,
			MonthlyFee:  u.MonthlyFee,
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

// AdminCreateUserRequest es el payload para crear un médico desde el panel admin.
// Todos los campos de perfil son obligatorios para cumplir NOM-024-SSA3-2012.
type AdminCreateUserRequest struct {
	Email             string `json:"email"`
	Password          string `json:"password"`
	NombreCompleto    string `json:"nombre_completo"`
	CedulaProfesional string `json:"cedula_profesional"`
	Especialidad      string `json:"especialidad"`
	Universidad       string `json:"universidad"`
	Direccion         string `json:"direccion"`
	Telefono          string `json:"telefono"`
	CURP              string `json:"curp,omitempty"`
	VendorRef         string `json:"vendor_ref,omitempty"`
}

// HandleAdminCreateUser crea un médico completo desde el panel admin:
// 1. Crea el usuario (credenciales + tenant)
// 2. Crea el perfil profesional (NOM-024)
// 3. Activa los módulos clínicos estándar
//
// POST /api/v1/admin/users
// Requiere AdminMiddleware.
func HandleAdminCreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}

	var req AdminCreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "payload invalido")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.NombreCompleto = strings.TrimSpace(req.NombreCompleto)
	req.CedulaProfesional = strings.TrimSpace(req.CedulaProfesional)
	req.Especialidad = strings.TrimSpace(req.Especialidad)
	req.Universidad = strings.TrimSpace(req.Universidad)
	req.Direccion = strings.TrimSpace(req.Direccion)
	req.Telefono = strings.TrimSpace(req.Telefono)

	switch {
	case req.Email == "":
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "email es obligatorio")
		return
	case req.Password == "":
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "password es obligatorio")
		return
	case len(req.Password) < 8:
		writeError(w, http.StatusBadRequest, "PASSWORD_TOO_SHORT", "password minimo 8 caracteres")
		return
	case req.NombreCompleto == "":
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "nombre_completo es obligatorio")
		return
	case req.Universidad == "":
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "universidad es obligatoria")
		return
	case req.Direccion == "":
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "direccion es obligatoria")
		return
	case req.Telefono == "":
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "telefono es obligatorio")
		return
	}

	// Validación NOM-024 delegada al shader mx_medical (ADR-0002, issue #203).
	// El handler es transporte puro; la regla de dominio vive en el Shader.
	if err := shaders.ValidateMxMedicalProfile(shaders.MxMedicalProfile{
		CedulaProfesional: req.CedulaProfesional,
		Especialidad:      req.Especialidad,
	}); err != nil {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", err.Error())
		return
	}

	if deps.UserRepo.ExistsByEmail(req.Email) {
		writeError(w, http.StatusConflict, "EMAIL_EXISTS", "el email ya esta registrado")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "HASH_ERROR", "error al procesar password")
		return
	}

	userID := "usr-" + strings.ReplaceAll(req.Email, "@", "-")
	tenantID := "tenant-" + userID
	curp := strings.ToUpper(strings.TrimSpace(req.CURP))

	u := postgres.User{
		CURP:         curp,
		ID:           userID,
		TenantID:     tenantID,
		Email:        req.Email,
		PasswordHash: hash,
		CreatedAt:    time.Now().UTC(),
	}
	if err := deps.UserRepo.Create(u); err != nil {
		if strings.Contains(err.Error(), "EMAIL_EXISTS") {
			writeError(w, http.StatusConflict, "EMAIL_EXISTS", "el email ya esta registrado")
			return
		}
		observability.Logger.Error("admin: error al crear usuario", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "DB_ERROR", "error al crear usuario")
		return
	}

	profile := ports.Profile{
		UserID:            userID,
		TenantID:          tenantID,
		Rubro:             "medico",
		NombreCompleto:    req.NombreCompleto,
		CedulaProfesional: req.CedulaProfesional,
		Especialidad:      req.Especialidad,
		Universidad:       req.Universidad,
		Direccion:         req.Direccion,
		Telefono:          req.Telefono,
	}
	if err := deps.ProfileRepo.Upsert(profile); err != nil {
		observability.Logger.Error("admin: error al crear perfil profesional",
			"user_id", userID, "error", err.Error())
		writeError(w, http.StatusInternalServerError, "PROFILE_ERROR",
			"usuario creado pero error al guardar perfil: "+err.Error())
		return
	}

	modulosEstandar := []string{"allergy", "prescription", "note"}
	modulosFallidos := []string{}
	for _, mod := range modulosEstandar {
		if err := deps.CapabilityRepo.Activate(tenantID, mod, "basico", 0); err != nil {
			observability.Logger.Error("admin: error al activar modulo",
				"tenant_id", tenantID, "module_id", mod, "error", err.Error())
			modulosFallidos = append(modulosFallidos, mod)
		}
	}

	// Asignar vendor_ref al tenant si se proporcionó (ADR-0026, issue #220).
	if v := strings.TrimSpace(req.VendorRef); v != "" && deps.VendorRepo != nil && deps.TenantRepo != nil {
		if _, err := deps.VendorRepo.GetByID(v); err == nil {
			// Vendor válido — asignar al tenant
			if err := deps.TenantRepo.SetVendorRef(tenantID, v); err != nil {
				observability.Logger.Error("admin: error al asignar vendor_ref",
					"tenant_id", tenantID, "vendor_ref", v, "error", err.Error())
			}
		}
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"data": map[string]any{
			"user_id":          userID,
			"tenant_id":        tenantID,
			"email":            req.Email,
			"modulos_activos":  modulosEstandar,
			"modulos_fallidos": modulosFallidos,
		},
		"error": nil,
	})
}

// HandleAdminUpdateProfile actualiza el perfil profesional de un médico desde el panel admin.
// PUT /api/v1/admin/users/:tenant_id/profile
func HandleAdminUpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/users/")
	tenantID = strings.TrimSuffix(tenantID, "/profile")
	if tenantID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "tenant_id requerido")
		return
	}

	var req struct {
		NombreCompleto    string `json:"nombre_completo"`
		CedulaProfesional string `json:"cedula_profesional"`
		Especialidad      string `json:"especialidad"`
		Universidad       string `json:"universidad"`
		Direccion         string `json:"direccion"`
		Telefono          string `json:"telefono"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "payload invalido")
		return
	}

	// Buscar el userID a partir del tenantID
	users, err := deps.UserRepo.FindAll()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al buscar usuario")
		return
	}
	var userID string
	for _, u := range users {
		if u.TenantID == tenantID {
			userID = u.ID
			break
		}
	}
	if userID == "" {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "tenant no encontrado")
		return
	}

	profile := ports.Profile{
		UserID:            userID,
		TenantID:          tenantID,
		Rubro:             "medico",
		NombreCompleto:    strings.TrimSpace(req.NombreCompleto),
		CedulaProfesional: strings.TrimSpace(req.CedulaProfesional),
		Especialidad:      strings.TrimSpace(req.Especialidad),
		Universidad:       strings.TrimSpace(req.Universidad),
		Direccion:         strings.TrimSpace(req.Direccion),
		Telefono:          strings.TrimSpace(req.Telefono),
	}
	if err := deps.ProfileRepo.Upsert(profile); err != nil {
		writeError(w, http.StatusInternalServerError, "PROFILE_ERROR", "error al actualizar perfil")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"ok": true}, "error": nil})
}

// HandleAdminSetBilling actualiza el modo de facturación de un tenant.
// PUT /api/v1/admin/users/:tenant_id/billing
func HandleAdminSetBilling(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/users/")
	tenantID = strings.TrimSuffix(tenantID, "/billing")
	if tenantID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "tenant_id requerido")
		return
	}

	var req struct {
		BillingMode string  `json:"billing_mode"`
		MonthlyFee  float64 `json:"monthly_fee"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "payload invalido")
		return
	}
	if req.BillingMode != "monthly" && req.BillingMode != "per_module" {
		writeError(w, http.StatusBadRequest, "INVALID_BILLING_MODE", "billing_mode debe ser 'monthly' o 'per_module'")
		return
	}

	if err := deps.UserRepo.SetBilling(tenantID, req.BillingMode, req.MonthlyFee); err != nil {
		writeError(w, http.StatusInternalServerError, "BILLING_ERROR", "error al actualizar facturacion")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"ok": true}, "error": nil})
}

// HandleAdminResetPassword resetea la contraseña de un médico desde el panel admin.
// PUT /api/v1/admin/users/:tenant_id/password
func HandleAdminResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/users/")
	tenantID = strings.TrimSuffix(tenantID, "/password")
	if tenantID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "tenant_id requerido")
		return
	}

	var req struct {
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "payload invalido")
		return
	}
	if len(req.NewPassword) < 8 {
		writeError(w, http.StatusBadRequest, "PASSWORD_TOO_SHORT", "password minimo 8 caracteres")
		return
	}

	// Buscar userID por tenantID
	users, err := deps.UserRepo.FindAll()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al buscar usuario")
		return
	}
	var userID string
	for _, u := range users {
		if u.TenantID == tenantID {
			userID = u.ID
			break
		}
	}
	if userID == "" {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "tenant no encontrado")
		return
	}

	hash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "HASH_ERROR", "error al procesar password")
		return
	}
	if err := deps.UserRepo.SetPassword(userID, hash); err != nil {
		writeError(w, http.StatusInternalServerError, "PASSWORD_ERROR", "error al actualizar contraseña")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"ok": true}, "error": nil})
}

// HandleAdminUpdateProfile actualiza el perfil profesional de un médico desde el panel admin.
// PUT /api/v1/admin/users/:tenant_id/profile
