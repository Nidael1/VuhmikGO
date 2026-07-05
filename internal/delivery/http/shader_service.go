package delivery

import (
	"fmt"
	"github.com/Nidael1/VuhmikGO/internal/application/ports"
	"github.com/Nidael1/VuhmikGO/internal/shaders"
)

// ShaderService es la única vía de comunicación entre la capa de entrega
// y los Shaders. Impide acceso directo al Core desde los handlers.
type ShaderService struct {
	medical  shaders.Shader
	export   shaders.ExportShader
	tenants  ports.TenantRepository
}

// NewShaderService construye el servicio con los Shaders autorizados.
func NewShaderService(tenants ports.TenantRepository) *ShaderService {
	return &ShaderService{
		medical:  shaders.NewMedicalBasicShader(), // fallback; Authorize resuelve dinámico
		export:   shaders.NewLegalExportShader(),  // fallback; Export resuelve dinámico
		tenants:  tenants,
	}
}

// Authorize evalúa si una operación está permitida para el tenant y actor dados.
// Retorna la decisión del Shader. El caller maneja el resultado.
func (s *ShaderService) Authorize(
	tenantID string,
	actorID string,
	op shaders.Operation,
) shaders.ShaderDecision {
	// Resolver shader dinámicamente por clinical_shader_key del tenant (ADR-0002, issue #204).
	// Fail-closed: si el tenant no existe en la tabla tenants, deniega.
	registry := shaders.NewShaderRegistry()
	medical := s.medical // fallback seguro
	if s.tenants != nil {
		if cfg, err := s.tenants.GetByID(tenantID); err == nil {
			medical = registry.Resolve(shaders.ShaderKey(cfg.ClinicalShaderKey))
		} else {
			return shaders.ShaderDecision{
				Result:    shaders.DecisionDeny,
				ErrorCode: "ER-SHADER-001",
				Reason:    "tenant no encontrado: " + tenantID,
			}
		}
	}
	ctx := shaders.ShaderContext{
		TenantID:  tenantID,
		ActorID:   actorID,
		Operation: op,
	}
	// Evaluar primero el clinical shader.
	decision := medical.Evaluate(ctx)
	if decision.Result != shaders.DecisionAllow {
		return decision
	}
	// Evaluar extra shaders activos (ADR-0025, issue #206).
	// Fail-closed: si algún extra shader deniega, la operación se deniega.
	if s.tenants != nil {
		if cfg, err := s.tenants.GetByID(tenantID); err == nil {
			extraRegistry := shaders.NewShaderRegistry()
			for _, key := range cfg.ExtraShaderKeys {
				extra := extraRegistry.Resolve(shaders.ShaderKey(key))
				if d := extra.Evaluate(ctx); d.Result != shaders.DecisionAllow {
					return d
				}
			}
		}
	}
	return decision
}

// Export genera un export legal en memoria vía LegalExportShader.
// El resultado debe usarse inmediatamente. No se persiste.
// No se registra PHI en logs.
func (s *ShaderService) Export(
	tenantID string,
	actorID string,
	evidenceID string,
) ([]byte, error) {
	// Resolver export shader dinámicamente por export_shader_key del tenant (ADR-0002, issue #205).
	// Fail-closed: key desconocido, vacío o export_none → deniega.
	exportShader := s.export // fallback seguro
	if s.tenants != nil {
		if cfg, err := s.tenants.GetByID(tenantID); err == nil {
			registry := shaders.NewExportShaderRegistry()
			if resolved := registry.Resolve(shaders.ExportShaderKey(cfg.ExportShaderKey)); resolved != nil {
				exportShader = resolved
			} else {
				return nil, fmt.Errorf("[ER-SHADER-003] export no autorizado para tenant: %s", tenantID)
			}
		} else {
			return nil, fmt.Errorf("[ER-SHADER-001] tenant no encontrado: %s", tenantID)
		}
	}
	ctx := shaders.ShaderContext{
		TenantID:   tenantID,
		ActorID:    actorID,
		Operation:  shaders.OperationExport,
		SubjectRef: evidenceID,
	}
	data := shaders.ExportData{
		EvidenceID: evidenceID,
		TenantID:   tenantID,
		State:      "issued",
	}
	return exportShader.GenerateExport(ctx, data)
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
