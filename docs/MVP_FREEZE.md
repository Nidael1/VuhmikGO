# MVP FREEZE — Cierre formal del MVP

## Fecha
2026-06-04

## Declaración

El MVP de VuhmikGO queda formalmente cerrado y congelado.
Issues #1 a #50 completados y mergeados a main.

A partir de este punto:
  - No se agregan nuevas tareas al alcance del MVP.
  - No se hacen commits adicionales bajo este freeze.
  - Cualquier cambio funcional posterior requiere un nuevo ciclo
    de planificación (post-MVP) con su propio ADR si aplica.

## Resumen de lo construido

Sprint 1 — ENGINE/Core (Issues #1-#15)
  Entidades, lifecycle, guards de inmutabilidad, void/replace,
  catalogos de error_code y reason_code, tests (9 PASS).

Sprint 2 — Shaders (Issues #16-#20)
  Contrato Shader, MedicalBasicShader, LegalExportShader,
  freeze auditado, tests (13 PASS).

Sprint 3 — CRM (Issues #21-#25, #31-#33)
  Frontend base, flujos CRUD via Shader, validaciones UX,
  manejo de errores tipados, freeze, auditoria de integracion,
  test manual happy path.

Sprint 4 — ECE (Issues #26-#30, #34-#40)
  Captura draft, emision+lock, void+replace, export legal
  eferimo, persistencia real PostgreSQL (pgx/v5).

Sprint 5 — Observabilidad (Issues #41-#45)
  Logging JSON sin PHI, metricas agregadas anonimas,
  validacion de secretos (fail-closed), control de acceso
  por contexto de tenant, runbooks operativos.

Sprint 6 (parcial) — Cierre (Issues #46-#49)
  Hardening, limpieza, alineacion de documentacion, demo E2E.

## Estado tecnico al freeze

  go build ./...  -> OK
  go vet ./...    -> OK
  gofmt -l .      -> limpio
  go test ./...   -> PASS (Core: 9, Shaders: 13)

## Garantias del MVP congelado

  - Core agnostico, determinista, append-only.
  - Acceso al Core unicamente via Shaders.
  - Lifecycle: draft -> issued -> locked/voided. Sin transiciones implicitas.
  - Inmutabilidad post-issued/locked (ER-CORE-001).
  - Void/Replace con reason_code obligatorio del catalogo.
  - Export legal efimero, sin persistencia, sin cache.
  - Observabilidad sin PHI/PII.
  - Fail-closed: secretos ausentes y contexto ausente bloquean operacion.
  - Multi-tenant via X-Tenant-ID/X-Actor-ID, validado por middleware.

## Documentos de freeze relacionados

  - docs/SHADER_FREEZE.md
  - docs/CRM_MVP_FREEZE.md
  - docs/ECE_MVP_FREEZE.md
  - docs/MVP_HARDENING.md
  - docs/CLEANUP_AUDIT.md
  - docs/SHADER_INTEGRATION_AUDIT.md
  - docs/TEST_MANUAL_CRM.md
  - docs/DEMO_E2E.md

## Regla final

Este documento marca el cierre formal del MVP. Issues #51 en
adelante (si se ejecutan) corresponden a preparacion de release
(versionado, tagging) y NO modifican el alcance funcional aqui
congelado.
