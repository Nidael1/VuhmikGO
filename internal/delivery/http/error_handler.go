package delivery

import (
	"net/http"

	"github.com/Nidael1/VuhmikGO/internal/shaders"
)

// renderShaderDeny renderiza una respuesta de denegación del Shader.
// Muestra el error_code explícitamente. No expone PHI.
// El CRM no interpreta la razón del error — solo la presenta.
func renderShaderDeny(w http.ResponseWriter, r *http.Request, d shaders.ShaderDecision) {
	w.WriteHeader(http.StatusForbidden)
	render(w, "layout.html", pageData{
		AppName:      appName(),
		PageTitle:    "Acceso denegado",
		ErrorCode:    d.ErrorCode,
		ErrorMessage: d.Reason,
	})
}

// renderUXError renderiza errores de validación UX.
// No contiene lógica Core ni business rules.
func renderUXError(w http.ResponseWriter, r *http.Request, errs []UXValidationError) {
	w.WriteHeader(http.StatusUnprocessableEntity)
	render(w, "layout.html", pageData{
		AppName:   appName(),
		PageTitle: "Error de validación",
		UXErrors:  errs,
	})
}
