package delivery

import (
	"fmt"
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
		SubjectRef: evidenceID,
	}
	data := shaders.ExportData{
		EvidenceID: evidenceID,
		TenantID:   tenantID,
		State:      "issued",
	}
	return s.export.GenerateExport(ctx, data)
}

// DraftResponse es el resultado del guardado de un borrador clínico.
// No expone entidades Core directamente.
type DraftResponse struct {
	ID      string `json:"id"`
	State   string `json:"state"`
	Message string `json:"message"`
}

// CreateDraft valida via Shader y crea un borrador clínico en memoria.
// El ID es un stub determinista — la persistencia real requiere el
// adaptador de repositorio (sprint posterior).
// No registra PHI en logs.
func (s *ShaderService) CreateDraft(
	tenantID string,
	actorID string,
	subjectID string,
) (DraftResponse, error) {
	ctx := shaders.ShaderContext{
		TenantID:  tenantID,
		ActorID:   actorID,
		Operation: shaders.OperationCreate,
		SubjectRef: subjectID,
	}
	decision := s.medical.Evaluate(ctx)
	if decision.Result != shaders.DecisionAllow {
		return DraftResponse{}, fmt.Errorf("[%s] %s", decision.ErrorCode, decision.Reason)
	}

	// Objeto draft creado en memoria — persistencia en sprint de repositorio
	stub := DraftResponse{
		ID:      "draft-" + subjectID,
		State:   "draft",
		Message: "borrador creado — persistencia pendiente de repositorio",
	}
	return stub, nil
}
