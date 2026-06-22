# ADR-0006 — UX fluida con versionado silencioso

## Estado
Aceptado

## Fecha
2026-06-21

## Contexto

VuhmikGO v1 está dirigido a médicos independientes en México que no
tienen cultura de sistemas de expediente clínico electrónico. La
fricción en la UX es el principal obstáculo de adopción.

Los sistemas de referencia (SS-MIX2 Japón, NHI Taiwán, EMR Corea)
comparten un patrón: UX simple para el médico, inmutabilidad y
trazabilidad completa en el backend, auditoría invisible al usuario.

México se encuentra en el momento equivalente al que tuvo Japón
antes de SS-MIX2: sin estándar establecido, regulación existente
pero no aplicada con rigor, y telemedicina con marco legal reciente
(Reforma enero 2026). Quien establezca el estándar técnico ahora
tiene ventaja de primer mover a escala nacional.

## Decision

La UX de VuhmikGO oculta la complejidad del versionado al médico.

El médico percibe que puede editar libremente sus registros.
Internamente el sistema ejecuta void + replace automáticamente,
preservando cada versión con su timestamp exacto.

El Core no cambia. La inmutabilidad es total. Solo cambia la capa
de presentación (Asteroides/Vue) y los endpoints de la API.

## Comportamiento visible para el médico

  - Crear nota → formulario simple, guardar
  - Editar nota → formulario pre-llenado, guardar (parece edición)
  - Historial → opcional, visible solo si el médico lo solicita
  - Sin menciones de void, replace, reason_code, lock, emit

## Comportamiento interno del sistema

  - Primera versión: draft → issued automático al guardar
  - Cada edición: void + replace silencioso con RC-REPLACE-002
  - El historial completo queda en la BD, trazable y auditable
  - COFEPRIS o auditor puede ver cada versión con su timestamp

## Protección legal

  Si en una demanda se requiere la nota del 15 de enero:
    - El sistema la tiene, inmutable, con timestamp exacto
    - El médico no puede alterarla retroactivamente
    - Aunque el médico crea que la editó, la versión original existe
    - La cadena de custodia es completa y verificable

## Impacto en el Core

  Ninguno. El Core permanece sin cambios.
  GuardMutation, GuardTransition, Void, Replace — sin modificar.

## Impacto en la API

  Se agrega: PUT /api/v1/evidence/:id
  Internamente ejecuta: void(original) + create(replacement) + emit
  Externamente retorna: el registro actualizado (sin exponer el void)

## Impacto en la UX Vue

  - Vista de edición simple (mismo formulario que captura)
  - Sin pantalla de confirmación de emisión
  - Sin selector de reason_code visible
  - El estado "emitido/bloqueado" no se muestra al médico
  - Historial de versiones disponible pero no prominente

## Referencia

  Patrón: SS-MIX2 (Japón), NHI MediCloud (Taiwán), EMR (Corea)
  Adaptación: versión mínima necesaria para v1 México
  Extensible a interoperabilidad nacional sin refactor del Core

## Consecuencias

  Positivas:
    - Adopción masiva por médicos sin cultura de sistemas
    - Cumplimiento legal completo sin fricción
    - Argumento sólido para UNAM/IPN como estándar
    - Defensible ante COFEPRIS en cualquier auditoría

  Negativas:
    - El médico no entiende que el sistema versiona
    - Requiere comunicación clara en términos de servicio
    - El historial invisible puede sorprender si se expone en auditoría
