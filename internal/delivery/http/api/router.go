package api

import (
	"net/http"

	"github.com/Nidael1/VuhmikGO/internal/auth"
)

// RegisterAPIRoutes registra las rutas de la API JSON /api/v1.
// Separadas de las rutas HTML historicas del MVP.
func RegisterAPIRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/auth/register", HandleRegister)
	mux.HandleFunc("/api/v1/auth/login", HandleLogin)
	mux.HandleFunc("/api/v1/auth/me", JWTMiddleware(HandleMe))
	mux.HandleFunc("/api/v1/evidence/draft", JWTMiddleware(HandleEvidenceDraft))
	mux.HandleFunc("/api/v1/evidence", JWTMiddleware(HandleEvidenceList))
	mux.HandleFunc("/api/v1/evidence/", JWTMiddleware(HandleEvidenceDetail))
	mux.HandleFunc("/api/v1/evidence/emit", JWTMiddleware(HandleEvidenceEmit))
	mux.HandleFunc("/api/v1/evidence/void", JWTMiddleware(HandleEvidenceVoid))
	mux.HandleFunc("/api/v1/evidence/replace", JWTMiddleware(HandleEvidenceReplace))
}

// JWTMiddleware protege un handler exigiendo un JWT valido.
// Extrae tenant_id y actor_id del token y los inyecta en el contexto.
func JWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if len(header) < 8 || header[:7] != "Bearer " {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "token requerido")
			return
		}
		claims, err := auth.ValidateToken(header[7:])
		if err != nil {
			writeError(w, http.StatusUnauthorized, "INVALID_TOKEN", "token invalido o expirado")
			return
		}
		r = r.WithContext(ContextWithClaims(r.Context(), claims))
		next(w, r)
	}
}
