# ESTADO SPRINT 7 — Post-MVP VuhmikGO

## Fecha
2026-06-22

## Contexto

El MVP v0.1.0-rc1 (Issues #1-#66) fue cerrado y archivado.
El Sprint 7 fue planificado en sesion de trabajo directa con el
producto owner y documentado mediante ADRs.

## ADRs generados en Sprint 7

  ADR-0004 — Frontend Vue SPA (docs/adr/ADR-0004-frontend-vue-spa.md)
  ADR-0005 — Autenticacion JWT propia (docs/adr/ADR-0005-jwt-auth.md)
  ADR-0006 — UX fluida con versionado silencioso
              (docs/adr/ADR-0006-fluid-ux-silent-versioning.md)

## Issues ejecutados Sprint 7.0 — Identidad y seguridad API

  #67 — ADR-0004 Frontend Vue SPA
  #68 — ADR-0005 JWT + Brand Book minimo
  #69 — Migracion tabla users (000003)
  #70 — Endpoints auth register/login/me + JWT
  #71 — JWT middleware context helpers

## Issues ejecutados Sprint 7.1 — API JSON backend Go

  #72 — Contrato REST JSON /api/v1
  #73 — Repository FindAll tenant-scoped
  #74 — GET /api/v1/evidence list + detail
  #75 — POST /api/v1/evidence/draft
  #76 — POST emit/void/replace endpoints
  #77 — POST /api/v1/evidence/:id/export
  #78 — Setup Vue 3 + Vite + TS + Router + Pinia
  #79 — Estructura Clean Architecture frontend
  #80 — Capa infrastructure HTTP client + repositories
  #81 — Layout sidebar Brand Book
  #82 — LoginView registro + login
  #83 — EvidenceListView
  #84 — EvidenceDraftView + EvidenceDetailView
  #85 — Vite proxy backend
  #86 — Fix middleware rutas publicas auth
  #87 — Fix middleware bypass /api/
  #88 — ADR-0006 UX fluida versionado silencioso
  #89 — HandleEvidenceEdit void+replace silencioso
  #90 — EvidenceDetailView + EvidenceEditView ADR-0006
  #91 — Fix UI bugs campo paciente + copy ADR-0006
  #92 — Fix router dynamic paths dispatcher
  #93 — Evidence content fields subject_id + notes
  #94 — Migracion tabla patients NOM-004 (000005)
  #95 — Patients API CRUD handlers
  #96 — Patients Vue list + new views
  #97 — Patient detail view con notas clinicas
  #98 — Draft view vinculada a patient query param
  #99 — Fix patients POST route dispatcher
  #100 — Fix navegacion sidebar IDs evidencia
  #101 — Fix login redirect subject_id notes draft emit
  #102 — Fix toItem subject_id notes fields
  #103 — Fix edit UpdateForVoid + patient name
  #104 — Patient notes history UX sin badges de estado

## Issues ejecutados Sprint 7.2 — UX adicional

  #105 — Expediente hoja continua sin badges

## Estado tecnico al corte

  go build ./...  -> OK
  go vet ./...    -> OK
  gofmt -l .      -> limpio
  go test ./...   -> 24 PASS
  frontend build  -> OK (Vite, sin errores TS)
  Migraciones aplicadas: 000001-000005

## Limitaciones conocidas (aceptadas para v1 demo)

  1. Store de usuarios y pacientes en memoria — se pierden al
     reiniciar el servidor. Requiere UserRepository y
     PatientRepository con adaptador PostgreSQL.

  2. CURP no implementado — acordado para iteracion posterior.
     La arquitectura lo soporta sin refactor (ALTER TABLE + campo).

  3. Store de evidencia en memoria — idem. El adaptador postgres
     existe pero no esta conectado al handler API.

  4. EvidenceDetailView aun accesible por URL directa — el flujo
     principal ya es via PatientDetailView.

## Pendiente prioritario para produccion real

  - Conectar adaptador PostgreSQL a los handlers de la API
    (actualmente usan inmemory).
  - Implementar UserRepository en PostgreSQL.
  - Implementar PatientRepository en PostgreSQL.
  - Tests del Sprint 7 (API + Vue).

## Regla

Este documento marca el estado del Sprint 7 al corte del
2026-06-22. Cualquier trabajo posterior debe referenciar este
documento como linea base.
