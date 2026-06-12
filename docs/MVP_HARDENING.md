# MVP HARDENING — Issue #46

## Fecha
2026-06-04

## Endpoints expuestos (post-hardening)

  GET/POST /                 — inicio
  GET      /dashboard         — dashboard
  GET      /pacientes         — pacientes
  GET/POST /ece/nuevo         — captura draft
  POST     /ece/draft/guardar — guardado draft via Shader
  GET/POST /ece/emitir        — emisión + lock
  GET/POST /ece/anular        — void + replace
  POST     /ece/exportar      — export legal efímero

## Endpoint eliminado

  /registros/nuevo (handleNuevoRegistro, crud_handlers.go)

Motivo: handler genérico de demostración del Issue #22, sin uso en
el MVP final. Su función fue reemplazada funcionalmente por los
endpoints /ece/* que sí persisten via repositorio.

## Guards de inmutabilidad activos

  GuardMutation     — internal/core/evidence/guard.go
  GuardContentEdit  — internal/core/evidence/guard.go
  GuardTransition   — internal/core/evidence/lifecycle.go

## Estado de tests

  internal/core/evidence — PASS (9 tests)
  internal/shaders       — PASS (13 tests)

## Debug / flags

  No existen flags de debug en el código. log/slog configurado en
  nivel Info por defecto (internal/observability/logger.go).

## Resultado

Build limpio, vet limpio, gofmt limpio, tests verdes.
Sin nuevas features. Sin cambios de contrato Core/Shader.
