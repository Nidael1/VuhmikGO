# ECE MVP FREEZE — Alcance del ECE congelado

## Estado
Activo desde: 2026-06-04

## Declaración

El alcance del ECE MVP queda congelado. No se agregan nuevas
funcionalidades al ECE hasta el siguiente sprint planificado.
Todo cambio requiere ADR aprobado.

## Flujos implementados (Issues #26–#29)

- Captura clínica en modo draft (/ece/nuevo)
- Emisión explícita y bloqueo post-emisión (/ece/emitir)
- Anulación y reemplazo con reason_code obligatorio (/ece/anular)
- Export legal bajo demanda sin persistencia (/ece/exportar)

## Garantías legales y clínicas del ECE MVP

- Solo se crea evidencia en estado draft
- La emisión es explícita — nunca automática
- Post-emisión: inmutabilidad total (issued + locked)
- Void requiere reason_code del catálogo RC-VOID-*
- El reemplazo genera nueva nota issued
- El historial original se preserva siempre (append-only)
- El export es efímero: en memoria, sin persistencia, sin cache
- No se registra PHI en logs en ningún flujo

## Auditoría de accesos directos al Core

Resultado: sin accesos directos al Core desde la capa ECE.

Comando: grep -r internal/core/evidence --include=*.go internal/delivery/ (vacío)

## Rutas ECE en freeze

- POST /ece/nuevo — captura draft
- GET  /ece/emitir — confirmación de emisión
- POST /ece/emitir — emisión + lock
- GET  /ece/anular — formulario void
- POST /ece/anular — void + replace
- POST /ece/exportar — export legal efímero

## Reglas del freeze

1. No se agregan rutas ECE sin ADR aprobado.
2. No se modifica el Core ni los Shaders.
3. No se relajan reglas de inmutabilidad ni de reason_code.
4. No se persisten archivos de export.
5. Toda nueva funcionalidad ECE requiere sprint planificado.

## Sprint siguiente: Sprint 5 — Observabilidad (Issues #31–#40)
