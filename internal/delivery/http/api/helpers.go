package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Nidael1/VuhmikGO/internal/auth"
)

// writeJSON escribe una respuesta JSON con el status dado.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError escribe una respuesta de error JSON estándar.
func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"data":  nil,
		"error": map[string]string{"code": code, "message": message},
	})
}

// claimsKey es la clave de contexto para los claims JWT.
type claimsKey struct{}

// ContextWithClaims agrega claims al contexto de la request.
func ContextWithClaims(ctx context.Context, claims *auth.Claims) context.Context {
	return context.WithValue(ctx, claimsKey{}, claims)
}
