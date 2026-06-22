package delivery

import (
	"net/http"
	"strings"

	"github.com/Nidael1/VuhmikGO/internal/shaders"
)

// publicPaths son las rutas que no requieren contexto de tenant/actor.
// Solo navegación de solo lectura sin operaciones sensibles.
var publicPaths = map[string]bool{
	"/api/v1/auth/register": true,
	"/api/v1/auth/login":    true,
	"/":          true,
	"/dashboard": true,
	"/pacientes": true,
}

// TenantContextMiddleware valida que cada request a rutas sensibles
// incluya X-Tenant-ID y X-Actor-ID.
//
// Reglas:
//   - Fail-closed: sin headers, la operación se rechaza (ER-SHADER-001).
//   - No accede al Core. Solo valida presencia de contexto.
//   - Las rutas en publicPaths quedan exentas (solo navegación).
//   - No contiene lógica de permisos — eso es responsabilidad del Shader.
func TenantContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if publicPaths[r.URL.Path] || strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}

		tenantID := r.Header.Get("X-Tenant-ID")
		actorID := r.Header.Get("X-Actor-ID")

		if tenantID == "" || actorID == "" {
			w.WriteHeader(http.StatusForbidden)
			render(w, "layout.html", pageData{
				AppName:      appName(),
				PageTitle:    "Acceso denegado",
				ErrorCode:    shaders.ErrShaderContextInvalid,
				ErrorMessage: UXCopyFor(shaders.ErrShaderContextInvalid, "contexto de tenant ausente"),
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}
