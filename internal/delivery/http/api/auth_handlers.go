package api

import (
	"crypto/sha256"
	"encoding/json"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Nidael1/VuhmikGO/internal/auth"
	"github.com/Nidael1/VuhmikGO/internal/infrastructure/postgres"
)

// RegisterRequest es el payload de registro.
type RegisterRequest struct {
	CURP     string `json:"curp"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest es el payload de login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse es la respuesta de autenticacion.
// Incluye access token (15min) y refresh token (7 dias).
type AuthResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	TenantID     string `json:"tenant_id"`
	ActorID      string `json:"actor_id"`
}

func issueTokenPair(user postgres.User) (AuthResponse, error) {
	// Access token — 15 minutos
	accessToken, err := auth.GenerateToken(user.ID, user.TenantID)
	if err != nil {
		return AuthResponse{}, fmt.Errorf("error al generar access token: %w", err)
	}

	// Refresh token — 7 dias, stateful en PostgreSQL
	plain, hash, err := postgres.GenerateRefreshTokenValue()
	if err != nil {
		return AuthResponse{}, fmt.Errorf("error al generar refresh token: %w", err)
	}

	rt := postgres.RefreshToken{
		ID:        "rt-" + user.ID + "-" + time.Now().Format("20060102150405"),
		UserID:    user.ID,
		TenantID:  user.TenantID,
		TokenHash: hash,
		ExpiresAt: time.Now().UTC().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now().UTC(),
	}
	if err := deps.RefreshTokenRepo.Create(rt); err != nil {
		return AuthResponse{}, fmt.Errorf("error al persistir refresh token: %w", err)
	}

	return AuthResponse{
		Token:        accessToken,
		RefreshToken: plain,
		TenantID:     user.TenantID,
		ActorID:      user.ID,
	}, nil
}

// HandleRegister registra un nuevo medico en el sistema.
//
// POST /api/v1/auth/register
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
	u := postgres.User{
		CURP:         strings.ToUpper(strings.TrimSpace(req.CURP)),
		ID:           userID,
		TenantID:     tenantID,
		Email:        req.Email,
		PasswordHash: hash,
		CreatedAt:    time.Now().UTC(),
	}
	if err := deps.UserRepo.Create(u); err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", "error al registrar usuario")
		return
	}
	resp, err := issueTokenPair(u)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "TOKEN_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": resp, "error": nil})
}

// HandleLogin autentica un medico y retorna access + refresh token.
//
// POST /api/v1/auth/login
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
	u, err := deps.UserRepo.FindByEmail(req.Email)
	if err != nil || !auth.CheckPassword(req.Password, u.PasswordHash) {
		writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "credenciales invalidas")
		return
	}
	resp, err := issueTokenPair(u)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "TOKEN_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": resp, "error": nil})
}

// HandleMe retorna el perfil del usuario autenticado.
//
// GET /api/v1/auth/me
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

// hashToken calcula SHA-256 de un token en texto plano.
func hashToken(plain string) string {
	h := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(h[:])
}

// RefreshRequest es el payload para renovar el access token.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// HandleRefresh renueva el access token usando un refresh token valido.
// El refresh token se rota — se revoca el anterior y se emite uno nuevo.
//
// POST /api/v1/auth/refresh
func HandleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "refresh_token es obligatorio")
		return
	}

	// Buscar el refresh token por hash
	hash := hashToken(req.RefreshToken)
	rt, err := deps.RefreshTokenRepo.FindByHash(hash)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "INVALID_REFRESH_TOKEN", "refresh token invalido o expirado")
		return
	}

	// Buscar el usuario
	u, err := deps.UserRepo.FindByID(rt.UserID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "USER_NOT_FOUND", "usuario no encontrado")
		return
	}

	// Revocar el refresh token actual (rotacion)
	if err := deps.RefreshTokenRepo.Revoke(rt.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al revocar token")
		return
	}

	// Emitir nuevo par de tokens
	resp, err := issueTokenPair(u)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "TOKEN_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": resp, "error": nil})
}

// LogoutRequest es el payload para cerrar sesion.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// HandleLogout revoca el refresh token activo cerrando la sesion.
// El access token expirara naturalmente en 15 minutos.
//
// POST /api/v1/auth/logout
func HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	var req LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "refresh_token es obligatorio")
		return
	}
	hash := hashToken(req.RefreshToken)
	rt, err := deps.RefreshTokenRepo.FindByHash(hash)
	if err != nil {
		// Si no existe o ya fue revocado, logout es exitoso igual
		writeJSON(w, http.StatusOK, map[string]any{"data": map[string]string{"message": "sesion cerrada"}, "error": nil})
		return
	}
	deps.RefreshTokenRepo.Revoke(rt.ID)
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]string{"message": "sesion cerrada"}, "error": nil})
}
