# GUIA DE DEMO EXTERNA — VuhmikGO MVP v0.1.0-rc1

## Fecha
2026-06-04

## Proposito

Guia para presentar el MVP a personas externas (potenciales
clientes, inversionistas, evaluadores tecnicos) de forma
controlada, sin datos reales y sin exponer informacion sensible.

## Audiencia

  - Tecnica: revisores de codigo, evaluadores de arquitectura
  - No tecnica: potenciales clientes, socios comerciales

## Entorno de despliegue controlado

Recomendado: entorno local o de staging aislado, nunca produccion
con datos reales.

  1. Clonar el repositorio en tag v0.1.0-rc1:

     git clone https://github.com/Nidael1/VuhmikGO.git
     cd VuhmikGO
     git checkout v0.1.0-rc1

  2. Base de datos aislada para demo:

     createdb vuhmik_demo
     export DATABASE_URL="postgres://localhost:5432/vuhmik_demo?sslmode=disable"
     migrate -path database/migrations -database "$DATABASE_URL" up

  3. Arrancar:

     go run ./cmd/vuhmik-api/

## Guion de presentacion (no tecnica)

  1. Mostrar pantalla de inicio (/) — branding white label configurable
  2. Navegar a "Nueva Nota Clinica" (/ece/nuevo) — explicar captura draft
  3. Mostrar pantalla de confirmacion de emision (/ece/emitir) —
     enfatizar la advertencia de irreversibilidad
  4. Explicar que tras emitir, el registro queda bloqueado
     permanentemente (cumplimiento legal/forense)
  5. Mostrar formulario de anulacion (/ece/anular) — explicar que
     "corregir" significa anular + crear nuevo registro, preservando
     el historial completo (auditoria)
  6. Mencionar el export legal bajo demanda, sin archivos guardados
     en servidor

## Guion tecnico (para revisores de codigo)

  1. ADRs en docs/adr/ — decisiones de arquitectura
  2. internal/core/evidence/ — Core agnostico, lifecycle, guards
  3. internal/shaders/ — unica frontera al Core
  4. docs/SHADER_FREEZE.md, CRM_MVP_FREEZE.md, ECE_MVP_FREEZE.md —
     auditorias de capas
  5. go test ./... — 22 tests, todos PASS
  6. docs/DEMO_E2E.md — flujo reproducible paso a paso

## Datos a usar en demo (ficticios, obligatorio)

  Tenant ID:  tenant-demo-001
  Actor ID:   actor-demo-001
  Subject ID: paciente-demo-001
  Notas: texto generico ("Nota clinica de ejemplo")

## Prohibiciones durante la demo

  - No usar nombres, expedientes ni datos de pacientes reales
  - No usar la base de datos de produccion
  - No modificar codigo durante la presentacion
  - No exponer DATABASE_URL ni variables de entorno en pantalla

## Cierre de la demo

Al finalizar, eliminar la base de datos de demo:

  dropdb vuhmik_demo

## Cambios de codigo en este issue

Ninguno. Documento puramente operativo.
