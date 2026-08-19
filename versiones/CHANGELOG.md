# VUHMÍK — Changelog

Registro acumulativo de lo que aporta cada issue/merge. Actualizar al cerrar cada issue.

---

## Sprint 9.5 — Corrección de bugs + Widget Agenda

### #163 — Widget "Agenda de hoy" en sidebar global
**Fecha:** 2026-08-19  
**Commit:** `e0e0c69`  
**Aporta:**
- Sidebar muestra las consultas del día en curso debajo de la navegación.
- Muestra hora, nombre corto del paciente y estado (✓ = emitida, opaco = en borrador).
- Contador `completadas/total` en el encabezado del widget.
- Se refresca automáticamente al navegar entre rutas (sin polling, usando `watch(route)`).
- Backend: `ConsultationItem` incluye `patient_nombre` enriquecido con lookup al `PatientRepo` en `HandleConsultationListAll`.
- Tipo TypeScript `Consultation` actualizado con `patient_nombre?: string`.

---

## Sprint 9.5 — Corrección de bugs de integridad en expediente

### #156 — Bugfix: PATCH parcial en edición inline de paciente
**Fecha:** 2026-08-19  
**Commits:** `4f3e3c9` (base, incluye cambios de tipo)  
**Aporta:**
- `saveName` y `saveSexo` en `PatientDetailView` ahora envían solo el campo que editan (PATCH parcial), eliminando la race condition que podía sobreescribir campos en DB silenciosamente.
- `PatientRequest` (TypeScript) convertido a campos opcionales.
- `saveSexo` reabre el selector al fallar la API para permitir reintentos.

### #157 — Bugfix: doble request al cambiar sexo (@change + @blur)
**Fecha:** 2026-08-19  
**Commit:** `4f3e3c9`  
**Aporta:**
- Elimina `@blur="saveSexo"` del `<select>` de sexo; solo `@change` dispara el guardado. Evitaba dos PATCHes concurrentes al seleccionar una opción.

### #158 — Bugfix: guard null en saveSexo antes de non-null assertion
**Fecha:** 2026-08-19  
**Commit:** `2651f51`  
**Aporta:**
- Captura `patient.value` en variable local antes de entrar al try. Si es null (race de navegación), sale limpiamente sin lanzar TypeError.

### #159 — Bugfix: mensaje de error incorrecto para fecha con formato inválido
**Fecha:** 2026-08-19  
**Commit:** `86a2dc9`  
**Aporta:**
- Separa la condición `err != nil` de `fn.After(...)` en `HandlePatientCreate` y `HandlePatientUpdate`. Fecha mal formateada ahora responde "debe tener formato YYYY-MM-DD" en lugar del mensaje engañoso "no puede ser futura".

### #160 — Bugfix: HandlePatientUpdate rechaza sexo inválido en lugar de ignorarlo
**Fecha:** 2026-08-19  
**Commit:** `0c8818c`  
**Aporta:**
- Si se envía `sexo` con valor distinto de M/F/I, el handler responde 400 `INVALID_FIELDS`. Antes lo ignoraba silenciosamente y devolvía 200 OK.

### #161 — Bugfix: fecha máxima del picker calculada en hora local, no UTC
**Fecha:** 2026-08-19  
**Commit:** `8b3a2cb`  
**Aporta:**
- Nueva función `localDateToday()` que construye la fecha de hoy con `getFullYear/getMonth/getDate` (hora local). Reemplaza `new Date().toISOString().slice(0,10)` que devolvía la fecha UTC y podía bloquear "hoy" para usuarios en zonas UTC+.

### #162 — Validación demográfica de paciente movida al Shader med_basic (ADR-0030)
**Fecha:** 2026-08-19  
**Commit:** `53b3de2`  
**Aporta:**
- Nueva función `ValidatePatientDemographics(nombre, fechaNacimiento, sexo)` en `internal/shaders/medical_basic.go` con código `ER-SHADER-010`.
- `HandlePatientCreate` y `HandlePatientUpdate` delegan toda validación clínica demográfica al Shader; los handlers solo deciden el código HTTP.
- ADR-0030 redactado y aceptado en `docs/adr/ADR-0030-validacion-demografica-shader.md`.

---

## Sprint 9.4 — Administración, Métricas y Consultas

### #155 — Módulo Consultas (ADR-0024)
**Commits:** `9615dca` → `e0d14c1` → `97e0e06`  
**Aporta:**
- Migración 000016: `consultation_projections` con signos vitales, `consultation_id` en `note_projections`
- `ConsultationShader`, `ConsultationService`, handlers REST y router
- Frontend: `ConsultationListView`, `ConsultationNewView` (formulario unificado signos vitales + nota + receta), sección en `PatientDetailView`, sidebar

### #154 — Signos Vitales en Notas (ADR-0022)
**Commits:** `9778d7a` → `d025eeb`  
**Aporta:**
- Migración 000015: signos vitales en `note_projections`, `clinical_note_id` en `prescription_projections`
- Fix CURP nullable (NULL en lugar de cadena vacía para evitar colisión en índice UNIQUE)
- Fix login/registro

### #153 — Campos Adicionales en Perfil Profesional (ADR-0021)
**Commits:** `92f3339` → `b8b2241`  
**Aporta:**
- Migración 000014: columnas `universidad`, `direccion`, `telefono` en `professional_profiles`

### #152 — Métricas y Actividad (ADR-0019 / ADR-0023)
**Commit:** `e092054`  
**Aporta:**
- Migración 000013: tabla `metrics_snapshot` (precálculo WAR-A) + `activity_log`

### #151 — Panel Admin (ADR-0018)
**Commit:** `380ba0c`  
**Aporta:**
- Frontend `AdminView`: lista de tenants, toggles de módulos, búsqueda, colapsable
- Redirección automática por rol (`is_admin` → `/admin`, médico → `/patients`)

### #150 — JWT con is_admin + Suspensión (ADR-0018)
**Commit:** `f860657`  
**Aporta:**
- Claims JWT incluyen `is_admin`
- `AdminMiddleware` para rutas `/admin/*`
- Login bloquea usuarios suspendidos

### #149 — Migración is_admin / is_suspended (ADR-0018)
**Commit:** `3a2ac11`  
**Aporta:**
- Migración 000012: columnas `is_admin`, `is_suspended` en `users`

---

## Sprint 9.3 — Receta Electrónica y CQRS

### #148 — Frontend Recetas UX
**Commit:** `e1c25df`  
**Aporta:**
- Nueva vista recetas con botones, orden, búsqueda, unificación de estilos

### #147 — Frontend Recetas Lista / Sección Paciente (ADR-0011)
**Commit:** `cd9b84b`  
**Aporta:**
- `PrescriptionListView` (lista global), sección recetas en `PatientDetailView`, emit desde frontend

### #146 — API Recetas (ADR-0011)
**Commit:** `3fe7453`  
**Aporta:**
- Handlers REST: draft, emit, void, list para recetas

### #145 — PrescriptionShader + CQRS (ADR-0011 / ADR-0022)
**Commit:** `5148e0e`  
**Aporta:**
- `PrescriptionShader`, `PrescriptionService`, proyecciones CQRS

### #144 — Refactor Notas a CQRS (ADR-0022)
**Commit:** `5296414`  
**Aporta:**
- `note_projections` migradas al patrón CQRS

### #143 — Refactor AllergyService CQRS (ADR-0022)
**Commit:** `1697705`  
**Aporta:**
- `AllergyService` refactorizado a CQRS, fix FK void+replace, orden consistente

### #142 — Migración Proyecciones CQRS (ADR-0022)
**Commit:** `2f9b8e0`  
**Aporta:**
- Migración 000011: `note_projections`, `allergy_projections`, `prescription_projections`

### #141 — Export Expediente Completo
**Commit:** `ef6aebd`  
**Aporta:**
- Export de expediente completo con hash de dos niveles
- Fix: de in-memory a postgres, fix filtro por `type`

### #140 — Pantalla Perfil Médico (ADR-0021)
**Commit:** `4977e83` → (anterior)  
**Aporta:**
- `ProfileView` (solo lectura), enlace en sidebar y footer

### #139 — API Perfil Profesional (ADR-0021)
**Commit:** `ad7528a` → `4977e83`  
**Aporta:**
- `profile_repository`, handlers `GET /profile`, `PUT /profile`

### #138 — Migración Perfil Profesional (ADR-0021)
**Commit:** `ad7528a`  
**Aporta:**
- Migración 000010: tabla `professional_profiles` (cédula, especialidad, rubro)

---

## Sprint 9.2b — Alergias y Edición Inline

### #137 — Editar / Anular Alergia Inline
**Commit:** `f6574ac`  
**Aporta:**
- Edición inline de alergias con void+replace silencioso desde `PatientDetailView`

### #136 — Edición Inline Nombre Paciente
**Commits:** `9142cd3` → `4a17035`  
**Aporta:**
- Ícono lápiz en header de paciente para editar nombre inline

---

## Sprint 9.2 — Capacidades, Alergias e IPS

### Fix post-sprint
**Commits:** `a4fa2c1` → `5530dfe` → `e86268a`  
**Aporta:**
- CURP opcional (null pointer), fix frontend (subject_ref, paths, blob), fix httpclient (no redirect en 401 en login)

### #135 — IPS Allergy Export (ADR-0010 / ADR-0012)
**Commit:** `afda43e`  
**Aporta:**
- `IPSAllergyExport`: proyección FHIR R4 de alergias para IPS

### #134 — Frontend Alergias (ADR-0012)
**Commit:** `21abaf5`  
**Aporta:**
- Sección alergias en `PatientDetailView` con barra de seguridad por criticidad

### #133 — API Alergias (ADR-0012)
**Commits:** `3228c5e` → `4297323`  
**Aporta:**
- Handlers `/patients/:id/allergies`, `/allergies/:id/void`, rutas en deps y main

### #132 — AllergyShader + CapabilityGuard (ADR-0012 / ADR-0017)
**Commit:** `dc86b73`  
**Aporta:**
- `AllergyShader`, `AllergyService`, integración con `CapabilityGuard`

### #131 — CapabilityGuard Fail-Closed (ADR-0017)
**Commit:** `54dd286`  
**Aporta:**
- `CapabilityGuard` como wrapper de Shaders: bloquea si módulo no activo

### #130 — CapabilityRepository (ADR-0017)
**Commit:** `e1a2d17`  
**Aporta:**
- Port `CapabilityRepository` e implementación postgres

### #129 — Migración Módulos + Tenant Capabilities (ADR-0017)
**Commit:** `56b1990`  
**Aporta:**
- Migración 000009: tablas `modules`, `tenant_capabilities` + seed médico

---

## Sprint 9.1 — Core Agnóstico

### #125 — Core Content Opaco + SubjectRef (ADR-0016)
**Commits:** `cc6d98c` → `c4bf987`  
**Aporta:**
- Core no conoce la clínica: content es blob opaco, `subject_ref` genérico
- Fix en `ece_draft_save_handler`

### #124 — Migración Core Content Opaco (ADR-0016)
**Commit:** `975cca1`  
**Aporta:**
- Migración 000008: content opaco + `subject_ref`

---

## Sprint 8.3 — WAR-A, Export y Seguridad

### #123 — CURP en Users y Patients
**Commit:** `aeb9be6`  
**Aporta:**
- Migración 000007, handlers y frontend para CURP

### #122 — Export XML (ADR-0007)
**Commit:** `6cd4cd7`  
**Aporta:**
- Export en XML HL7 CDA + JSON con selector de formato (`Accept` header)

### #121 — Hash SHA-256 en Export (ADR-0008)
**Commit:** `3d95503`  
**Aporta:**
- Hash SHA-256 en evidencia exportada para integridad probatoria

### #120 — Checklist GoLive
**Commit:** `2972b20`  
**Aporta:**
- Verificación de estado completo del sistema

### #119 — MetricsPurgeWorker (WAR-A)
**Commit:** `f6a02fa`  
**Aporta:**
- Worker Go: purga métricas con retención 30 días

### #118 — BackupWorker (WAR-A)
**Commit:** `63e7c5a`  
**Aporta:**
- Worker Go: backup PostgreSQL cada 24h, purge a 7 días

### #117 — Redis Integration (WAR-A)
**Commit:** `24115be`  
**Aporta:**
- Redis como broker para workers WAR-A

---

## Sprint 8.2 — Auth Stateful

### #116 — Logout + Revocación Tokens
**Commit:** `d7c43c8`  
**Aporta:**
- Endpoint logout, revocación de refresh token en Redis

### #115 — Refresh Endpoint + Rotación
**Commit:** `e5f44eb`  
**Aporta:**
- `POST /auth/refresh` con rotación de token

### #114 — JWT 15min + Refresh Tokens (ADR stateful)
**Commit:** `52cf3e4`  
**Aporta:**
- JWT con expiración 15 min, refresh tokens con expiración 7 días

### #113 — Migración Refresh Tokens
**Commit:** `3e1eb8a`  
**Aporta:**
- Migración 000006: tabla `refresh_tokens`

---

## Sprint 7.2 — Frontend Vue + Pacientes + Expediente

### #112 a #109 — Lista de Pacientes
**Aporta:** Sort A-Z, búsqueda, contador, fix fecha PostgreSQL, repositorio postgres de pacientes

### #108 — UserRepository Postgres
**Aporta:** Adaptador postgres para usuarios y auth handlers

### #107 — ADR-0007/0008/0009
**Aporta:** Export XML+JSON, firma, traspaso de expediente

### #105 — Expediente Hoja Continua
**Aporta:** Vista de expediente sin badges sin estado, hoja continua

### #104 — Patient Notes History UX
**Aporta:** Historial de notas de paciente, UX sin state badges

### #103 / #102 / #101 / #100 / #99 — Fixes nav, routes, ids
**Aporta:** Fix redirect login, subject_id en notas, navigation con IDs, routes dispatcher

### #98 — Draft View con Patient Query Param
### #97 — PatientDetailView con notas clínicas
### #96 — Patients Vue: lista, vistas, router
### #95 — Patients API CRUD
### #94 — Migración Patients (NOM-004)
### #93 — Evidence content fields + subject_id
### #92 — Fix dynamic path routing
### #91 / #90 / #89 — Fluid UX Edit (ADR-0006 void+replace silencioso)
### #87 / #86 / #85 — Fix API middleware, auth routes, Vite proxy
### #84 / #83 / #82 / #81 — Vue views: evidence draft/detail/list, login, layout sidebar
### #80 / #79 / #78 — Vue infrastructure: http client, repositories, estructura, router, Pinia, Vite+TS

---

## Sprint 7.1 — API REST Evidence

### #77 — Evidence Export Endpoint
### #76 — Evidence Emit / Void / Replace Endpoints
### #75 — Evidence Draft Endpoint
### #74 — Evidence List y Detail
### #73 — Repository FindAll tenant-scoped
### #72 — API Contract REST JSON
### #71 — JWT Middleware + Context Helpers
### #70 — Auth Register / Login / Me
### #69 — Migración Users
### #68 — ADR-0005: JWT Auth + Brand Book mínimo
### #67 — ADR-0004: Frontend Vue SPA

---

## Sprint 6.4 — MVP Freeze

Issues #50–#66: architecture freeze, contract freeze, schema freeze, behavior freeze, tenant isolation review, non-persistent export, error code review, internal audit, final approval, demo materials, release candidate, release tag, mvp archive.

**Capacidad entregada:** sistema congelado y auditado listo para sprint 7.

---

## Sprints 1–5 — Core ENGINE

| Issue | Sprint | Aporte |
|-------|--------|--------|
| #1 | 1.1 | Core entities base |
| #2 | 1.2 | Core state constants (draft/issued/locked/voided) |
| #3 | 1.1 | Core relations |
| #4 | 1.2 | Migraciones forward-only |
| #5 | 1.2 | Índices críticos del core |
| #6 | 1.2 | Core immutability guards |
| #7 | 1.3 | Lifecycle transitions matrix |
| #8 | 1.3 | Block edit post-issue |
| #9 | 1.3 | Void evidence |
| #10 | 1.3 | Replace evidence |
| #11 | 1.4 | Error code catalog |
| #12 | 1.4 | Error code integration |
| #13 | 1.4 | Reason code catalog |
| #14 | 1.4 | Reason code enforcement |
| #15 | 1.5 | Minimal core tests |
| #16 | 2.1 | Shader interface |
| #17 | 2.2 | MedicalBasicShader |
| #18 | 2.3 | LegalExportShader |
| #19 | 2.4 | Shader boundary tests |
| #20 | 2.4 | Shader freeze |
| #21 | 3.1 | CRM base frontend |
| #22 | 3.2 | CRM CRUD flows |
| #23 | 3.3 | CRM UX validations |
| #24 | 3.4 | CRM error handling |
| #25 | 3.4 | CRM MVP freeze |
| #26 | 4.1 | ECE draft capture |
| #27 | 4.2 | ECE issue + lock |
| #28 | 4.3 | ECE void + replace |
| #29 | 4.4 | ECE legal export |
| #30 | 4.4 | ECE MVP freeze |
| #31 | 3.4 | CRM visual error handling |
| #32 | 3.4 | CRM-Shader integration |
| #33 | 3.4 | CRM manual happy path test |
| #34 | 4.1 | ECE draft capture UI |
| #35 | 4.1 | ECE draft save via Shader |
| #36 | 4.1 | ECE draft capture (refactor) |
| #37 | 4.2 | ECE issue endpoint |
| #38 | 4.2 | ECE lock issued |
| #39 | 4.3 | ECE void |
| #40 | 4.3 | ECE void+replace |
| #41 | 5.1 | Structured logging (JSON) |
| #42 | 5.2 | Aggregated metrics |
| #43 | 5.3 | Runtime secrets |
| #44 | 5.3 | Runtime access control |
| #45 | 5.4 | Runbooks |
| #46 | 6.1 | MVP hardening |
| #47 | 6.1 | Cleanup + go.mod |
| #48 | 6.2 | Docs alignment |
| #49 | 6.3 | E2E demo |

---

## Cómo actualizar este archivo

Al cerrar cada issue, agregar una entrada al sprint activo con:

```
### #N — Título del issue
**Commit:** `hash`
**Aporta:**
- Qué capacidad nueva tiene el sistema
- Qué migración se aplicó (si aplica)
- Qué ADR implementa (si aplica)
```
