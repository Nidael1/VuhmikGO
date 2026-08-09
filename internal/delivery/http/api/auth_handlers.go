package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/Nidael1/VuhmikGO/internal/auth"
	"github.com/Nidael1/VuhmikGO/internal/infrastructure/postgres"
	"github.com/Nidael1/VuhmikGO/internal/observability"
)

// Configuracion de rate limiting para login.
// Se bloquea la IP tras 5 intentos fallidos en 15 minutos.
// El contador vive en Redis con TTL automatico.
const (
	loginMaxAttempts = 5
	loginWindowTTL   = 15 * time.Minute
)

// loginRateLimitKey retorna la clave Redis para el contador de intentos.
func loginRateLimitKey(ip string) string {
	return "login_attempts:" + ip
}

// registerRequest es el payload de registro.
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
	Token         string `json:"token"`
	RefreshToken  string `json:"refresh_token"`
	TenantID      string `json:"tenant_id"`
	ActorID       string `json:"actor_id"`
	IsAdmin       bool   `json:"is_admin"`
	TermsAccepted bool   `json:"terms_accepted"`
}

func issueTokenPair(user postgres.User) (AuthResponse, error) {
	// Access token — 15 minutos
	accessToken, err := auth.GenerateToken(user.ID, user.TenantID, user.IsAdmin)
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
		Token:         accessToken,
		RefreshToken:  plain,
		TenantID:      user.TenantID,
		ActorID:       user.ID,
		IsAdmin:       user.IsAdmin,
		TermsAccepted: user.TermsAcceptedAt != nil && user.TermsVersion != nil && *user.TermsVersion == CurrentTermsVersion,
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
		observability.Logger.Error("error al crear usuario", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
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
	// Rate limiting por IP: bloquear antes de tocar la base de datos.
	// Usar IP como clave evita enumerar emails validos por diferencia de tiempo.
	// Extraer IP sin puerto. RemoteAddr tiene formato "ip:puerto".
	ip := r.RemoteAddr
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		ip = strings.SplitN(fwd, ",", 2)[0]
	}
	ip = strings.TrimSpace(ip)
	if host, _, err := net.SplitHostPort(ip); err == nil {
		ip = host
	}
	rdb := deps.RedisClient.RDB()
	rateLimitKey := loginRateLimitKey(ip)

	attempts, _ := rdb.Get(context.Background(), rateLimitKey).Int()
	if attempts >= loginMaxAttempts {
		observability.Logger.Warn("login bloqueado por rate limit", "ip", ip, "attempts", attempts)
		writeError(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "demasiados intentos, intenta en 15 minutos")
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
		// Incrementar contador con TTL. INCR no resetea el TTL existente,
		// asi que solo se setea el TTL en el primer intento.
		pipe := rdb.Pipeline()
		pipe.Incr(context.Background(), rateLimitKey)
		pipe.Expire(context.Background(), rateLimitKey, loginWindowTTL)
		if _, err := pipe.Exec(context.Background()); err != nil {
			observability.Logger.Error("error al incrementar rate limit", "error", err.Error())
		}
		go func() {
			id := fmt.Sprintf("la-%d", time.Now().UnixNano())
			_, _ = deps.DB.Exec(context.Background(),
				`INSERT INTO login_attempts (id, email, occurred_at, reason) VALUES ($1, $2, NOW(), 'invalid_credentials')`,
				id, req.Email,
			)
		}()
		writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "credenciales invalidas")
		return
	}
	if u.IsSuspended {
		writeError(w, http.StatusForbidden, "ACCOUNT_SUSPENDED", "cuenta suspendida, contacta al administrador")
		return
	}
	resp, err := issueTokenPair(u)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "TOKEN_ERROR", err.Error())
		return
	}
	// Login exitoso: resetear contador de intentos fallidos.
	rdb.Del(context.Background(), rateLimitKey)
	logActivity(r.Context(), u.TenantID, "session_start")
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

// AcceptTermsRequest es el payload para registrar aceptacion de terminos.
type AcceptTermsRequest struct {
	Version string `json:"version"`
}

// HandleAcceptTerms registra la aceptación de términos y condiciones del usuario autenticado.
//
// El cliente debe enviar la version del documento que efectivamente mostro.
// Si no coincide con CurrentTermsVersion se rechaza: registrar aceptacion de
// una version distinta a la exhibida invalidaria el valor probatorio.
//
// POST /api/v1/auth/accept-terms
func HandleAcceptTerms(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	claims, ok := r.Context().Value(claimsKey{}).(*auth.Claims)
	if !ok || claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}
	var req AcceptTermsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PAYLOAD", "payload invalido")
		return
	}
	if req.Version != CurrentTermsVersion {
		writeError(w, http.StatusConflict, "TERMS_VERSION_MISMATCH", "version de terminos no vigente")
		return
	}
	if err := deps.UserRepo.AcceptTerms(claims.ActorID, req.Version); err != nil {
		observability.Logger.Error("error al registrar terminos", "error", err.Error())
		writeError(w, http.StatusInternalServerError, "DB_ERROR", "error al registrar aceptacion")
		return
	}
	logActivity(r.Context(), claims.TenantID, "terms_accepted")
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]string{"message": "terminos aceptados"}, "error": nil})
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
	logActivity(r.Context(), rt.TenantID, "session_end")
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]string{"message": "sesion cerrada"}, "error": nil})
}
