# ADR-0004 — Frontend Vue SPA post-MVP

## Estado
Aceptado

## Fecha
2026-06-12

## Contexto

El MVP VuhmikGO v0.1.0-rc1 fue cerrado con una capa de presentacion
HTML renderizada desde Go (internal/delivery/http/templates/).
Para el ciclo post-MVP se requiere una interfaz web moderna, separada
del backend, apta para presentacion a VC y uso real por medicos.

## Decision

Se adopta Vue 3 + Vite + TypeScript + Vue Router + Pinia como stack
frontend. La SPA vivira en /frontend dentro del mismo monorepo.
La comunicacion con el backend sera REST JSON bajo /api/v1.
Los templates HTML historicos se conservan sin modificacion.

## Stack

  Vue 3          — framework UI
  Vite           — build tool y dev server
  TypeScript     — tipado de contratos y componentes
  Vue Router     — navegacion SPA
  Pinia          — estado global (sesion, auth, datos compartidos)

## Comunicacion

  REST JSON bajo /api/v1
  No GraphQL en esta etapa

## Ubicacion

  Monorepo — carpeta /frontend

## Estructura interna frontend

  frontend/src/
    app/            — router, stores, configuracion
    domain/         — modelos, tipos TypeScript
    application/    — casos de uso frontend
    infrastructure/ — cliente HTTP, repositorios API
    presentation/   — views, components, layouts

## Impacto en backend

  Los handlers HTML actuales se conservan.
  Se agregan handlers JSON nuevos bajo internal/delivery/http/api/
  Los servicios de aplicacion existentes se reutilizan.

## Alternativas consideradas

  Mantener HTML templates — rechazado: no escalable para VC ni produccion
  GraphQL — rechazado: complejidad innecesaria en esta etapa
  Repo separado — rechazado: producto todavia evoluciona de forma conjunta

## Consecuencias

  Se agrega tooling Node/Vite al proyecto.
  Se requiere build frontend separado.
  Se debe mantener compatibilidad entre tipos TypeScript y Go.
