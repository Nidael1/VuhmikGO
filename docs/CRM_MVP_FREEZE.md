# CRM MVP FREEZE — Alcance del CRM congelado

## Estado
Activo desde: 2026-06-04

## Declaración

El alcance del CRM MVP queda congelado. No se agregan nuevas
funcionalidades al CRM hasta el siguiente sprint planificado.
Todo cambio fuera del roadmap requiere ADR aprobado.

## Funcionalidades implementadas (Issues #21–#24)

- Layout base y routing mínimo (/, /dashboard, /pacientes, /registros/nuevo)
- Flujos CRUD administrativos no clínicos vía Shaders
- Validaciones UX de formato y campos requeridos
- Manejo de errores tipados con error_code visible

## Lo que el CRM NO hace (por diseño)

- No accede al Core directamente
- No contiene lógica clínica ni reglas de negocio
- No persiste datos (la persistencia es responsabilidad del Core)
- No interpreta ni redefine errores del Core o Shaders
- No expone PHI en respuestas ni en logs

## Auditoría de accesos directos al Core

Resultado: sin accesos directos al Core desde la capa de entrega.

Comando ejecutado:
grep -r internal/core/evidence --include=*.go internal/delivery/ (resultado vacío)

## Stack del CRM en freeze

- Protocolo: net/http + ServeMux
- Templates: html/template (server-rendered)
- Shaders: MedicalBasicShader (via ShaderService)
- Validaciones: UXValidator (solo formato y required)
- Errores: renderShaderDeny + renderUXError (error_code visible)

## Reglas del freeze

1. No se agregan rutas ni handlers sin ADR aprobado.
2. No se modifica el contrato de ShaderService sin ADR.
3. No se accede al Core directamente desde delivery.
4. No se agrega lógica de negocio en handlers.
5. Toda nueva funcionalidad CRM requiere un sprint planificado.

## Sprint siguiente: Sprint 4 — PRODUCT_ECE (Issues #26–#30)
