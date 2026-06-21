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

// decisionFromError convierte un error genérico en un ShaderDecision de deny.
// Usado cuando el Export retorna error sin ShaderDecision explícita.
func decisionFromError(err error) shaders.ShaderDecision {
	return shaders.ShaderDecision{
		Result:    shaders.DecisionDeny,
		ErrorCode: shaders.ErrShaderContextInvalid,
		Reason:    err.Error(),
	}
}
