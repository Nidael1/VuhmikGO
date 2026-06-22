package api

import (
	"net/http"
	"strings"

	"github.com/Nidael1/VuhmikGO/internal/auth"
)

// RegisterAPIRoutes registra las rutas de la API JSON /api/v1.
func RegisterAPIRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/auth/register", HandleRegister)
	mux.HandleFunc("/api/v1/auth/login", HandleLogin)
	mux.HandleFunc("/api/v1/auth/me", JWTMiddleware(HandleMe))
	mux.HandleFunc("/api/v1/evidence", JWTMiddleware(HandleEvidenceList))
	mux.HandleFunc("/api/v1/evidence/draft", JWTMiddleware(HandleEvidenceDraft))

	// Dispatcher dinamico para rutas con ID variable:
	// /api/v1/evidence/:id
	// /api/v1/evidence/:id/emit
	// /api/v1/evidence/:id/void
	// /api/v1/evidence/:id/replace
	// /api/v1/evidence/:id/export
	// /api/v1/evidence/:id/edit
	mux.HandleFunc("/api/v1/evidence/", JWTMiddleware(evidenceDispatcher))
}

// evidenceDispatcher enruta requests con ID dinamico en el path.
// ServeMux de Go no soporta path params — este dispatcher los resuelve.
func evidenceDispatcher(w http.ResponseWriter, r *http.Request) {
	// Elimina el prefijo base
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/evidence/")
	// path puede ser: "abc123", "abc123/emit", "abc123/export", etc.
	parts := strings.SplitN(path, "/", 2)

	if len(parts) == 1 {
		// GET /api/v1/evidence/:id
		HandleEvidenceDetail(w, r)
		return
	}

	action := parts[1]
	switch action {
	case "emit":
		HandleEvidenceEmit(w, r)
	case "void":
		HandleEvidenceVoid(w, r)
	case "replace":
		HandleEvidenceReplace(w, r)
	case "export":
		HandleEvidenceExport(w, r)
	case "edit":
		HandleEvidenceEdit(w, r)
	default:
		writeError(w, http.StatusNotFound, "NOT_FOUND", "ruta no encontrada")
	}
}

// JWTMiddleware protege un handler exigiendo un JWT valido.
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
