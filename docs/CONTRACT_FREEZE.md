# CONTRACT FREEZE — Issue #62

## Fecha
2026-06-12

## Version congelada
v0.1.0-rc1

## Declaracion

Los contratos entre Core y Shaders, y entre Shaders y la capa de
aplicacion, quedan formalmente congelados. No se modifican DTOs,
interfaces ni firmas de metodos sin un nuevo ciclo de planificacion.

## Contratos congelados

### Contrato Core (internal/core/evidence/)

  Evidence struct:
    ID, TenantID, State, CreatedAt, IssuedAt*, VoidedAt*, ReplacedByID*

  State: draft | issued | locked | voided

  Operaciones de dominio:
    GuardMutation(Evidence) error
    GuardContentEdit(Evidence) error
    GuardTransition(State, State) error
    Void(Evidence, ReasonCode, time.Time) (Evidence, error)
    Replace(Evidence, Evidence, ReasonCode, time.Time) (Evidence, Evidence, error)

  Codigos de error: ER-CORE-001, 002, 003, 004

  Codigos de razon: RC-VOID-001..004, RC-REPLACE-001..002

### Contrato Shader (internal/shaders/)

  ShaderContext: TenantID, Operation, SubjectID, ActorID, Country
  ShaderDecision: Result (allow|deny), ErrorCode, Reason
  ExportData: EvidenceID, TenantID, State, CreatedAt, IssuedAt*, VoidedAt*, ReplacedByID*

  interface Shader: Evaluate(ShaderContext) ShaderDecision
  interface ExportShader: Shader + GenerateExport(ShaderContext, ExportData) ([]byte, error)

  Operaciones: create | void | replace | read | export
  Codigos de error: ER-SHADER-001 (ErrShaderContextInvalid), ER-SHADER-002 (ErrShaderOperationDenied)

### Contrato Repository (internal/application/ports/)

  interface EvidenceRepository:
    Create(Evidence) error
    FindByID(tenantID, id string) (Evidence, error)
    Update(tenantID string, Evidence) error

## Reglas del freeze

1. No se agregan campos a Evidence sin migracion y ADR.
2. No se cambia la firma de Evaluate, GenerateExport ni los metodos
   del repositorio sin ADR.
3. No se crean nuevos codigos ER-CORE-* ni ER-SHADER-* sin ADR.
4. No se agregan operaciones al catalogo de Operations sin ADR.
5. Los DTOs ShaderContext y ShaderDecision son inmutables.
