package shaders

import "github.com/Nidael1/VuhmikGO/internal/application/ports"

// CapabilityGuard envuelve cualquier Shader y agrega la compuerta
// fail-closed de capacidades (ADR-0017) antes de evaluar.
//
// Flujo:
//  1. Verifica que el modulo esta publicado (plano de control).
//  2. Verifica que el tenant tiene el modulo activo (plano de datos).
//  3. Si ambas pasan, delega al Shader interno.
//
// Si el repositorio falla, deniega (fail-closed).
type CapabilityGuard struct {
	inner    Shader
	caps     ports.CapabilityRepository
	moduleID string
	rubro    string
}

// NewCapabilityGuard crea un Shader con compuerta fail-closed.
func NewCapabilityGuard(
	inner Shader,
	caps ports.CapabilityRepository,
	moduleID string,
	rubro string,
) *CapabilityGuard {
	return &CapabilityGuard{
		inner:    inner,
		caps:     caps,
		moduleID: moduleID,
		rubro:    rubro,
	}
}

// Evaluate verifica capacidades antes de delegar al Shader interno.
func (g *CapabilityGuard) Evaluate(ctx ShaderContext) ShaderDecision {
	// 1. Publicado en el catalogo (plano de control)
	published, err := g.caps.IsPublished(g.moduleID, g.rubro)
	if err != nil || published == false {
		return ShaderDecision{
			Result:    DecisionDeny,
			ErrorCode: ErrShaderOperationDenied,
			Reason:    "modulo no publicado para el rubro: " + g.moduleID,
		}
	}

	// 2. Activo para el tenant (plano de datos)
	active, err := g.caps.IsActive(ctx.TenantID, g.moduleID)
	if err != nil || active == false {
		return ShaderDecision{
			Result:    DecisionDeny,
			ErrorCode: ErrShaderOperationDenied,
			Reason:    "modulo no activo para el tenant: " + g.moduleID,
		}
	}

	// 3. Delegar al Shader interno
	return g.inner.Evaluate(ctx)
}
