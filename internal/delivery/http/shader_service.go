package delivery

import (
	"github.com/Nidael1/VuhmikGO/internal/shaders"
)

// ShaderService es la única vía de comunicación entre la capa de entrega
// y los Shaders. Impide acceso directo al Core desde los handlers.
type ShaderService struct {
	medical shaders.Shader
	export  shaders.ExportShader
}

// NewShaderService construye el servicio con los Shaders autorizados.
func NewShaderService() *ShaderService {
	return &ShaderService{
		medical: shaders.NewMedicalBasicShader(),
		export:  shaders.NewLegalExportShader(),
	}
}

// Authorize evalúa si una operación está permitida para el tenant y actor dados.
// Retorna la decisión del Shader. El caller maneja el resultado.
func (s *ShaderService) Authorize(
	tenantID string,
	actorID string,
	op shaders.Operation,
) shaders.ShaderDecision {
	ctx := shaders.ShaderContext{
		TenantID:  tenantID,
		ActorID:   actorID,
		Operation: op,
	}
	return s.medical.Evaluate(ctx)
}

// Export genera un export legal en memoria vía LegalExportShader.
// El resultado debe usarse inmediatamente. No se persiste.
// No se registra PHI en logs.
func (s *ShaderService) Export(
	tenantID string,
	actorID string,
	evidenceID string,
) ([]byte, error) {
	ctx := shaders.ShaderContext{
		TenantID:  tenantID,
		ActorID:   actorID,
		Operation: shaders.OperationExport,
		SubjectID: evidenceID,
	}
	data := shaders.ExportData{
		EvidenceID: evidenceID,
		TenantID:   tenantID,
		State:      "issued",
	}
	return s.export.GenerateExport(ctx, data)
}
