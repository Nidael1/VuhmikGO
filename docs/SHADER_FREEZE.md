# SHADER FREEZE — Capa de Shaders congelada

## Estado
Activo desde: 2026-06-04

## Declaración

La capa de Shaders queda congelada como única vía de acceso al Core.
Toda operación sobre evidencia Core debe pasar por un Shader.
Ninguna capa exterior puede importar tipos de internal/core/evidence directamente.

## Auditoría de accesos (Issue #20)

Resultado: sin accesos directos al Core fuera de las capas autorizadas.
Comando ejecutado: grep -r internal/core/evidence --include=*.go . (resultado vacío)

## Capas autorizadas para acceder al Core

- internal/core/evidence/ — es el Core mismo
- internal/shaders/ — única frontera contractual
- Tests unitarios del Core y Shaders — autorizados

## Capas NO autorizadas para acceder al Core directamente

- internal/application/ — accede vía Shaders únicamente
- internal/delivery/ — accede vía Shaders únicamente
- Asteroides (UI/producto) — accede vía Shaders únicamente

## Shaders disponibles post-freeze

- MedicalBasicShader — perfil med_basic
- LegalExportShader — perfil legal_export

## Reglas del freeze

1. No se agregan Shaders nuevos sin ADR aprobado.
2. No se modifica el contrato Shader ni sus DTOs sin ADR aprobado.
3. No se accede al Core directamente desde ninguna capa exterior.
4. La auditoría de accesos debe repetirse antes de cada merge al Sprint 3+.
5. Cualquier violación requiere ADR antes de ejecutar.

## Referencias

- internal/shaders/shader.go — contrato canónico
- ADR-0001, ADR-0002, ADR-0003
