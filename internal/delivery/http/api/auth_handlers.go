// Package api implementa los handlers JSON de la API /api/v1.
// No renderiza HTML. Retorna exclusivamente JSON.
// No accede al Core directamente — usa servicios de aplicacion.
package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Nidael1/VuhmikGO/internal/auth"
)

// User representa un usuario del sistema en memoria para el MVP.
// La persistencia real requiere UserRepository (issue posterior).
type userRecord struct {
	ID           string
	TenantID     string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

// userStore es un store en memoria para demo.
// Sera reemplazado por repositorio PostgreSQL en issue posterior.
var userStore = map[string]*userRecord{}

// RegisterRequest es el payload de registro.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest es el payload de login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse es la respuesta de autenticacion.
type AuthResponse struct {
	Token    string `json:"token"`
	TenantID string `json:"tenant_id"`
	ActorID  string `json:"actor_id"`
}

// ErrorResponse es el formato estandar de error de la API.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"data":  nil,
		"error": ErrorResponse{Code: code, Message: message},
	})
}

// HandleRegister registra un nuevo medico en el sistema.
//
// POST /api/v1/auth/register
// Body: {"email": "...", "password": "..."}
func HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "payload invalido")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "email y password son obligatorios")
		return
	}
	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "PASSWORD_TOO_SHORT", "password minimo 8 caracteres")
		return
	}
	if _, exists := userStore[req.Email]; exists {
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
	userStore[req.Email] = &userRecord{
		ID:           userID,
		TenantID:     tenantID,
		Email:        req.Email,
		PasswordHash: hash,
		CreatedAt:    time.Now().UTC(),
	}
	token, err := auth.GenerateToken(userID, tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "TOKEN_ERROR", "error al generar token")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"data":  AuthResponse{Token: token, TenantID: tenantID, ActorID: userID},
		"error": nil,
	})
}

// HandleLogin autentica un medico y retorna un JWT.
//
// POST /api/v1/auth/login
// Body: {"email": "...", "password": "..."}
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "payload invalido")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	user, exists := userStore[req.Email]
	if !exists || !auth.CheckPassword(req.Password, user.PasswordHash) {
		writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "credenciales invalidas")
		return
	}
	token, err := auth.GenerateToken(user.ID, user.TenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "TOKEN_ERROR", "error al generar token")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data":  AuthResponse{Token: token, TenantID: user.TenantID, ActorID: user.ID},
		"error": nil,
	})
}

// HandleMe retorna el perfil del usuario autenticado.
//
// GET /api/v1/auth/me
// Header: Authorization: Bearer <token>
func HandleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	claims, ok := r.Context().Value(claimsKey{}).(*auth.Claims)
	if !ok || claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]string{
			"actor_id":  claims.ActorID,
			"tenant_id": claims.TenantID,
		},
		"error": nil,
	})
}

// claimsKey es la clave de contexto para los claims JWT.
type claimsKey struct{}

// ContextWithClaims agrega claims al contexto de la request.
func ContextWithClaims(ctx context.Context, claims *auth.Claims) context.Context {
	return context.WithValue(ctx, claimsKey{}, claims)
}
