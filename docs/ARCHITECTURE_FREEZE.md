# ARCHITECTURE FREEZE — Issue #61

## Fecha
2026-06-12

## Version congelada
v0.1.0-rc1

## Declaracion

La arquitectura del MVP VuhmikGO queda formalmente congelada.
No se permiten cambios estructurales posteriores sin un nuevo
ciclo de planificacion con ADR aprobado.

## Arquitectura congelada

Monolito modular hexagonal sobre un unico binario Go.

Capas (en orden de dependencia):

  ENGINE/Core (internal/core/evidence/)
    - Entidades de dominio puras
    - Reglas de lifecycle y guards de inmutabilidad
    - Sin dependencias externas a Go stdlib
    - Agnostico: sin UI, sin pais, sin reglas clinicas especificas

  Shaders (internal/shaders/)
    - Unica frontera contractual de acceso al Core
    - DTOs: ShaderContext, ShaderDecision, ExportData
    - Implementaciones: MedicalBasicShader, LegalExportShader
    - Sin logica clinica ni administrativa

  Application (internal/application/)
    - Casos de uso: ECEService
    - Puertos: EvidenceRepository (interface)
    - Orquesta Core + repositorio

  Infrastructure (internal/infrastructure/)
    - Adaptadores: postgres (pgx/v5), inmemory (tests)
    - SQL explicito, sin ORM

  Delivery (internal/delivery/http/)
    - net/http + ServeMux
    - html/template (server-rendered)
    - Middleware de contexto de tenant

  Observability (internal/observability/)
    - log/slog JSON
    - Metricas atomicas en memoria
    - Validacion de secretos

## ADRs que rigen esta arquitectura

  ADR-0001 — Stack canonico Go
  ADR-0002 — Sistema de Shaders por pais y modo generico
  ADR-0003 — Estructura Go idiomatica, monolito modular hexagonal

## Reglas del freeze

1. No se agregan nuevas capas sin ADR.
2. No se modifica el contrato entre capas sin ADR.
3. No se introduce ORM, framework web ni dependencias de UI.
4. No se divide en microservicios ni multi-repo.
5. El Core permanece agnostico permanentemente.
