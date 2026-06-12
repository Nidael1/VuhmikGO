# RELEASE NOTES — v0.1.0-rc1

## Fecha
2026-06-04

## Tipo
Release Candidate 1 del MVP (Issues #1-#50)

## Resumen

Primera version etiquetable de VuhmikGO. Implementa el MVP completo
segun el plan de ejecucion: Core agnostico, Shaders como frontera
contractual, CRM administrativo, ECE clinico con persistencia
PostgreSQL, y capa de observabilidad.

## Componentes incluidos

  ENGINE/Core
    - Entidad Evidence (7 campos), lifecycle draft->issued->locked/voided
    - Guards de inmutabilidad (ER-CORE-001/002/003/004)
    - Void y Replace con reason_code obligatorio
    - 9 tests unitarios (PASS)

  Shaders
    - Contrato Shader (DTOs ShaderContext/ShaderDecision)
    - MedicalBasicShader (perfil med_basic)
    - LegalExportShader (export efimero en memoria)
    - 13 tests de frontera (PASS)

  CRM (Asteroide administrativo)
    - Layout, routing, validaciones UX
    - Manejo de errores tipados (error_code visible)

  ECE (Asteroide clinico)
    - Captura draft, emision+lock, void+replace, export legal
    - Persistencia real via PostgreSQL (pgx/v5)

  Observabilidad
    - Logging JSON sin PHI (log/slog)
    - Metricas agregadas anonimas
    - Validacion de secretos fail-closed
    - Middleware de contexto de tenant

## Stack tecnico

  Go 1.25, PostgreSQL (pgx/v5), migraciones forward-only
  (golang-migrate), net/http + ServeMux, html/template

## Requisitos para correr

  Ver docs/runbooks/ARRANQUE.md

## Limitaciones conocidas de este RC

  - Sin autenticacion real (headers X-Tenant-ID/X-Actor-ID son
    de confianza del cliente; auth real es post-MVP)
  - Sin telemedicina (roadmap futuro, ver ADR-0002)
  - Sin shaders de pais especificos (mx_medical reservado, no activo)
  - Templates HTML minimos, sin diseño de marca final

## Cambios funcionales en este issue

Ninguno. Este documento es preparacion de release.
