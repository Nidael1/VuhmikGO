package api

import (
	"net/http"

	"github.com/Nidael1/VuhmikGO/internal/auth"
)

// TenantIDFromContext extrae el tenant_id del contexto JWT.
// Retorna vacio si no hay claims — el JWTMiddleware ya bloqueo antes.
func TenantIDFromContext(r *http.Request) string {
	claims, ok := r.Context().Value(claimsKey{}).(*auth.Claims)
	if !ok || claims == nil {
		return ""
	}
	return claims.TenantID
}

// ActorIDFromContext extrae el actor_id del contexto JWT.
func ActorIDFromContext(r *http.Request) string {
	claims, ok := r.Context().Value(claimsKey{}).(*auth.Claims)
	if !ok || claims == nil {
		return ""
	}
	return claims.ActorID
}
