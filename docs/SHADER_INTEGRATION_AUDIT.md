# SHADER INTEGRATION AUDIT — CRM → Shader

## Fecha
2026-06-04

## Resultado
APROBADO — sin accesos directos al Core desde la capa de entrega.

## Auditoría ejecutada (Issue #32)

### Accesos directos al Core desde delivery
Comando: grep -r "internal/core" --include="*.go" internal/delivery/
Resultado: vacío (cero ocurrencias)

### Punto único de integración CRM → Shader
Todos los handlers de datos usan NewShaderService() como única vía:

- crud_handlers.go      → NewShaderService().Authorize(...)
- ece_handlers.go       → NewShaderService().Authorize(...)
- ece_issue_handlers.go → NewShaderService().Authorize(...)
- ece_void_handlers.go  → NewShaderService().Authorize(...)
- ece_export_handlers.go → NewShaderService().Export(...)

### Punto único de integración (ShaderService)
Definido en: internal/delivery/http/shader_service.go
Métodos disponibles:
- Authorize(tenantID, actorID, operation) ShaderDecision
- Export(tenantID, actorID, evidenceID) ([]byte, error)

## Garantías

1. Ningún handler accede a internal/core/evidence directamente.
2. Toda operación de datos pasa por ShaderService.
3. ShaderService es el único punto que instancia Shaders.
4. Errores tipados (error_code) preservados en todos los flujos.
5. No se agregaron features nuevas en este issue.

## Regla de enforcement

Si en cualquier PR futuro aparece un import de internal/core
desde internal/delivery, el PR debe ser rechazado sin excepción.
