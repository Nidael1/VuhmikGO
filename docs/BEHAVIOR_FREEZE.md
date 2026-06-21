# BEHAVIOR FREEZE — Issue #64

## Fecha
2026-06-12

## Version congelada
v0.1.0-rc1

## Declaracion

El comportamiento funcional observable del MVP queda formalmente
congelado. No se cambia la logica de ningun flujo sin ADR aprobado.

## Flujos congelados

### Flujo 1 — Captura draft

  POST /ece/nuevo o POST /ece/draft/guardar
  Input: subject_id, notes, X-Tenant-ID, X-Actor-ID
  Comportamiento:
    1. Middleware valida X-Tenant-ID y X-Actor-ID (fail-closed)
    2. UXValidator valida presencia y formato
    3. ShaderService.Authorize(tenantID, actorID, create)
    4. MedicalBasicShader.Evaluate -> allow/deny
    5. Si allow: registro creado en state=draft
  Output: HTTP 201 con DraftResponse (id, state, message)

### Flujo 2 — Emision y bloqueo

  POST /ece/emitir
  Comportamiento:
    draft -> issued (IssuedAt asignado)
    issued -> locked (inmutable permanente)
  Garantia: no reversible. GuardTransition bloquea retroceso.

### Flujo 3 — Anulacion y reemplazo

  POST /ece/anular
  Input: reason_code (catalogo RC-VOID-*), replacement_notes
  Comportamiento:
    1. reason_code validado contra ReasonCodeCatalog
    2. Original: state=voided, voided_at asignado
    3. Reemplazo: nuevo registro state=issued, replaced_by_id enlazado
    4. Historial original preservado (sin DELETE ni UPDATE de contenido)
  Garantia: el original nunca desaparece.

### Flujo 4 — Export legal

  POST /ece/exportar
  Comportamiento:
    1. LegalExportShader.Evaluate -> allow si OperationExport
    2. GenerateExport -> json.Marshal en memoria
    3. Response: Content-Disposition: attachment, Cache-Control: no-store
    4. Sin escritura a disco en ningun punto
  Garantia: efimero. No persiste tras la respuesta HTTP.

### Flujo 5 — Rechazo por contexto ausente

  Cualquier ruta sensible sin X-Tenant-ID o X-Actor-ID
  Comportamiento:
    Middleware retorna HTTP 403 con error_code ER-SHADER-001
    antes de llegar al handler.
  Garantia: fail-closed estricto.

## Comportamientos que NO cambian sin ADR

  - La matriz de transiciones de lifecycle
  - El catalogo de ReasonCode
  - El catalogo de error_code Core y Shader
  - El comportamiento de GuardMutation (bloquea issued/locked)
  - El formato de ExportData (JSON en memoria)
  - El filtrado por tenant_id en repositorio
