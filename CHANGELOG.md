# CHANGELOG — VUHMÍK

Registro de issues ejecutados sobre el Asteroide `crm_ui`. Cada entrada corresponde a un issue independiente: una rama, un commit, un PR.

---

## Estado agregado por ADR (última actualización: issue #237, 2026-07-09)

Esta tabla es un tablero de estado, no un registro cronológico. Se actualiza cada vez
que un ADR cambia de estado real (no solo cuando se documenta). Fuente: `docs/adr/*.md`
más verificación directa contra código en `main`.

| ADR | Tema | Estado | Nota |
|---|---|---|---|
| 0001 | Stack Go | ✅ Implementado | — |
| 0002 | Shaders por país / modo genérico | ✅ Implementado | — |
| 0003 | Estructura Go idiomática | ✅ Implementado | — |
| 0004 | Frontend Vue SPA | ✅ Implementado | — |
| 0005 | JWT auth | ✅ Implementado | — |
| 0006 | UX fluida / versionado silencioso | ✅ Implementado | — |
| 0007 | Export clínico XML+JSON | ⚠️ Parcial | Cubierto en la práctica por IPS/FHIR (ADR-0010); no hay implementación dedicada al formato original propuesto |
| 0008 | Firma digital / integridad | ❌ No implementado | Sin ningún archivo de código correspondiente |
| 0009 | Traspaso de paciente entre tenants | ✅ Implementado | issues #221, #222 |
| 0010 | IPS/FHIR como modelo de intercambio | ✅ Implementado | issue #212 |
| 0011 | Medicación y receta electrónica | ✅ Implementado | Backend completo; advertencia de controlados bloqueada por decisión regulatoria pendiente (no es gap de código) |
| 0012 | Alergias e intolerancias | ✅ Implementado | issues #132, #133, #135 |
| 0013 | Diagnósticos CIE-10 | ✅ Implementado | Backend issue #235; UI issue #237 — validado end-to-end 2026-07-09 |
| 0014 | Inmunizaciones | ✅ Implementado | Backend issue #235; UI issue #237 — validado end-to-end 2026-07-09 |
| 0015 | Resultados de laboratorio | ✅ Implementado | Backend issue #235; UI issue #237 — validado end-to-end 2026-07-09 |
| 0016 | Core agnóstico (content opaco) | ✅ Implementado | — |
| 0017 | Registro de capacidades por tenant | ✅ Implementado | CapabilityGuard conectado en los 6 módulos publicados tras issues #235 y #236 |
| 0018 | Panel de toggles por cuenta | ✅ Implementado | Validado en navegador 2026-07-09 |
| 0019 | Panel de métricas de negocio | ✅ Implementado | Validado en navegador 2026-07-09; bugs de columna corregidos |
| 0020 | Super-admin off-web | ⏸️ Diferido (decisión, no gap) | v1 = seed de `modules` por migración |
| 0021 | Perfil profesional por rubro | ✅ Implementado | — |
| 0022 | CQRS proyecciones de lectura | ✅ Implementado | Proyecciones conectadas para los 6 módulos tras issue #235 |
| 0023 | Panel de actividad y uso | ✅ Implementado | Validado en navegador 2026-07-09; migración 000026 aplicada |
| 0024 | Consulta médica | ✅ Implementado | — |
| 0025 | Modelo de datos de Tenant | ✅ Implementado | — |
| 0026 | Referencia de vendedor en tenant | ✅ Implementado | — |
| 0027 | Audit Package ZIP síncrono | ✅ Implementado | — |
| 0028 | Import de IPS Bundle FHIR R4 externo | ✅ Implementado | — |

### ✅ Validado end-to-end en navegador (2026-07-09)
- Notas clínicas: Create + Edit + panel lateral
- Alergias e intolerancias: Create + Edit + Void
- Diagnósticos CIE-10: Create + Void
- Inmunizaciones: Create + Void
- Resultados de laboratorio: Create + Void
- Consultas médicas: Create + detalle
- Recetas electrónicas: Backend OK — requiere perfil profesional completo en tenant
- Panel admin — Operaciones: lista, toggles, crear médico
- Panel admin — Métricas: snapshot real con datos
- Panel admin — Actividad: tabla lista, se poblará con uso

### 🔧 Pendiente de código real
- **ADR-0008 (firma digital / integridad):** sin ningún avance.
- **ADR-0007 (export XML+JSON):** su alcance original quedó parcialmente
  absorbido por IPS/FHIR (ADR-0010); requiere decisión explícita.
- **Advertencia de controlados (ADR-0011):** bloqueada por decisión regulatoria —
  requiere lista oficial COFEPRIS Grupo II/III validada antes de implementar.

### 🐛 Bugs conocidos (no bloqueantes)
- UX: secciones clínicas no se auto-expanden al guardar un registro nuevo.
- UX: panel de notas generales no es visible hasta expandir manualmente al abrir formulario.
- UX: `activity_snapshot` no registra conteos de diagnósticos/inmunizaciones/lab_results
  (solo nota, alergia, receta, exportación, paciente, sesión).

### Hallazgo colateral (no ADR)
Falta la migración `000025` en la secuencia numérica (existe `000024` y `000026`).
Coincide con el caso ya conocido de una rama de `allergy_projections` creada,
diagnosticada y borrada sin merge.

---

## Issue A — Jerarquía visual de secciones en expediente del paciente

**Rama:** `issue/crm-patient-detail-visual-hierarchy`
**Commit:** `1fcb0e4`
**Capa:** Asteroide `crm_ui` (UX/Producto). No toca Shaders ni Core.

### Problema
En la vista de expediente del paciente, las secciones Alergias, Recetas electrónicas y Consultas tenían el mismo peso visual (mismo color de header, mismo estilo de card), sin jerarquía que permitiera distinguir entre dato de fondo (alergias), prescripción médica (recetas) y actividad clínica cronológica (consultas).

### Solución
- Cada sección recibió un acento visual propio mediante borde izquierdo y color de fondo de header:
  - Alergias: acento naranja (seguridad/advertencia)
  - Recetas electrónicas: acento azul clínico (prescripción)
  - Consultas: acento turquesa (color principal de marca, actividad cronológica)
- Se agregó icono y contador de registros a cada header de sección.
- Las recetas pasaron de lista de texto plano a tarjetas (`rx-card`) con badge de estado "emitida".
- Se reemplazaron los emojis (⚠ 💊 🩺) por iconos SVG lineales monocromos, consistentes con el Brand Book (sin glow, sin gradientes, trazo limpio) — requerido por tratarse de software clínico para hospitales/consultorios.
- Se incrementó el espaciado entre secciones para mejorar la respiración visual.

### Archivos involucrados
- `frontend/src/presentation/views/PatientDetailView.vue` — único archivo modificado.

### Fuera de alcance (no tocado)
- Lógica de negocio, llamadas a repositorios, modelo de datos.
- Navegación de `PrescriptionListView.vue` y `ConsultationListView.vue` (corresponde a Issue B).

---

## Issue B1 — Backend: endpoint de detalle de receta individual

**Rama:** `issue/156-prescription-detail-endpoint`
**Commit:** `c8a72e4`
**Merge a main:** `61880cc`
**Capa:** Aplicación/Infraestructura (servicio + repositorio + handler). No toca Core ni Shaders — sigue el mismo patrón ya existente para consultas.

### Problema
`PrescriptionProjectionRepository` no exponía un método para obtener una receta individual por su ID. Solo existían `ListByPatient` y `ListAll`. Esto impedía construir una vista de detalle de receta sin cargar listas completas, y rompía la consistencia con `ConsultationProjectionRepository`, que ya tenía `FindByID`.

### Solución
Se replicó exactamente el patrón ya usado en consultas, en cinco puntos:
- `ports.PrescriptionProjectionRepository` — se agregó `FindByID(tenantID, evidenceID string) (PrescriptionProjection, error)` a la interfaz.
- `postgres.PrescriptionProjectionRepository` — se implementó `FindByID` con el mismo query/scan que `ListAll`, filtrando por `evidence_id` y `tenant_id`.
- `PrescriptionService` — se agregó el método `FindByID` que delega al repositorio.
- `prescription_handlers.go` — se agregó `HandlePrescriptionDetail` (`GET /api/v1/prescriptions/:id`).
- `router.go` — se actualizó `prescriptionDispatcher` para reconocer `GET /:id` como caso de detalle, igual que `consultationDispatcher`.

### Archivos involucrados
- `internal/application/ports/prescription_projection_repository.go`
- `internal/application/prescription_service.go`
- `internal/delivery/http/api/prescription_handlers.go`
- `internal/delivery/http/api/router.go`
- `internal/infrastructure/postgres/prescription_projection_repository.go`

### Verificación
Build de Go limpio (`go build ./...`) antes del commit.

---

## Issue B2 — Frontend: vistas de detalle de receta y consulta, navegación corregida

**Rama:** `issue/157-prescription-consultation-detail-view-v2`
**Commit:** `4de06df`
**Merge a main:** `cb64d18`
**Capa:** Asteroide `crm_ui` (UX/Producto). Depende de Issue B1 ya mergeado.

### Problema
Al abrir una receta o consulta desde su sección propia en el menú lateral (`Recetas`, `Consultas`), el sistema redirigía al perfil completo del paciente (`/patients/{id}`) en lugar de mostrar el detalle de esa instancia específica.

### Solución
Se decidió que cada receta y consulta debe tener su propia página de detalle (no modal), por tratarse de documentos clínico-legales que requieren URL propia, persistente y compartible (consistente con el lifecycle draft→issued→locked/voided y el export bajo demanda).

- `prescriptionRepository.ts` — se agregó `get(id)` consumiendo el nuevo endpoint de Issue B1.
- `PrescriptionDetailView.vue` (nueva) — vista de solo lectura: medicamento, dosis, diagnóstico, indicaciones, seguimiento, fecha de emisión, link al paciente.
- `ConsultationDetailView.vue` (nueva) — vista de solo lectura: signos vitales, nota clínica vinculada (vía `evidenceRepository`, mismo patrón que `PatientDetailView.vue`, mostrando "sin nota" cuando no existe), link al paciente.
- `router/index.ts` — se agregaron las rutas `/prescriptions/:id` y `/consultations/:id`.
- `PrescriptionListView.vue` — el click en una receta ahora navega a `/prescriptions/{id}` en vez de `/patients/{id}`.
- `ConsultationListView.vue` — el `RouterLink` de cada consulta ahora apunta a `/consultations/{id}` en vez de `/patients/{id}`.

### Archivos involucrados
- `frontend/src/infrastructure/repositories/prescriptionRepository.ts`
- `frontend/src/presentation/views/PrescriptionDetailView.vue` (nuevo)
- `frontend/src/presentation/views/ConsultationDetailView.vue` (nuevo)
- `frontend/src/router/index.ts`
- `frontend/src/presentation/views/PrescriptionListView.vue`
- `frontend/src/presentation/views/ConsultationListView.vue`

### Verificación
Build de frontend limpio (`npm run build`, incluye type-check). Probado en navegador: navegación desde lista, refresh (F5) directo sobre la URL de detalle, nota clínica visible.

---

## Fix post-merge — referencia a campo inexistente bloqueaba build

**Commit:** `2051d8f` (directo en `main`)

### Problema
Al mergear Issue A a `main`, el archivo `PatientDetailView.vue` traía una referencia a `rx.consultation_id` dentro del bloque de "receta vinculada a consulta". Ese campo no existe en el tipo `Prescription` actual — pertenece a trabajo pendiente de otro issue (ver Notas de ejecución), y solo estaba presente porque Issue A se trabajó sobre una copia de respaldo del repositorio que tenía ese trabajo mezclado sin commitear. El error no aparecía en `npm run dev` (sin type-check estricto) pero sí bloqueaba `npm run build`.

### Solución
Se removió el bloque de markup que dependía de `rx.consultation_id` (el chip de "receta vinculada" dentro de cada entrada de consulta en el expediente del paciente). Esta funcionalidad se reincorporará cuando se trabaje formalmente el issue que vincula receta↔consulta.

### Archivos involucrados
- `frontend/src/presentation/views/PatientDetailView.vue`

---

## Notas de ejecución

Durante esta sesión se identificó una desincronización entre dos copias locales del repositorio:
- `/Volumes/D/vuhmikGo` — repositorio real, conectado a `origin` (GitHub: `Nidael1/VuhmikGO`).
- `/Volumes/D/Copia de VuhmikGO` — copia de respaldo sin `.git` propio, usada temporalmente como espacio de trabajo durante Issue A.

Se identificó trabajo de backend pendiente sin commitear (prescripción vinculada a `consultation_id`, migración `database/migrations/000017_prescription_consultation_id.up.sql`, y el archivo `internal/delivery/http/api/prescription_print.go`). Este trabajo se resguardó de dos formas:

```
git stash push -m "wip-issue-156-prescription-consultation-link"
```

Y el archivo `prescription_print.go` (que no estaba trackeado por git) se apartó temporalmente en `/tmp/pending_issue156_real_2/` para no bloquear los builds de los issues de esta sesión, sin perderlo.

Este trabajo pendiente corresponde, siguiendo la numeración interna de commits del proyecto, al **issue #158 o posterior** (siguiente disponible tras el cierre de #155, #156 y #157 en esta sesión). Queda pendiente de asignarse formalmente como su propio issue antes de retomarse — incluye: agregar `consultation_id` a `Prescription`/`PrescriptionProjection`, aplicar la migración `000017`, y reincorporar el endpoint de impresión de receta (`prescription_print.go`).

---

## Issue C1 (#158) — Backend: vincular receta a consulta (consultation_id)

**Rama:** `issue/158-prescription-consultation-link`
**Commit:** `24fb881`
**Merge a main:** `3b0773f`
**Capa:** Aplicación/Infraestructura. No toca Core ni Shaders.

### Problema
La receta no tenía vínculo con la consulta que la originó. Sin ese vínculo, el endpoint de impresión no podía recuperar los signos vitales de la consulta para incluirlos en el PDF.

### Solución
- Migración `000017_prescription_consultation_id.up.sql` — agrega columna `consultation_id TEXT` e índice a `prescription_projections` (ya aplicada en DB antes del issue, formalizada en git aquí).
- `ports.PrescriptionProjection` — se agregó campo `ConsultationID string`.
- `postgres.PrescriptionProjectionRepository` — se actualizaron `Upsert`, `ListAll`, `ListByPatient` y `FindByID` para incluir la columna. Se usó `COALESCE(consultation_id, '')` para manejar NULLs históricos.
- `PrescriptionService.CreateDraft` — recibe `consultationID string` y lo persiste en la proyección. `Emit` preserva el valor leyendo la proyección existente antes de sobreescribir.
- `prescription_handlers.go` — `PrescriptionRequest` y `PrescriptionItem` incluyen `consultation_id`. El handler pasa el valor al servicio.

### Archivos involucrados
- `database/migrations/000017_prescription_consultation_id.up.sql`
- `internal/application/ports/prescription_projection_repository.go`
- `internal/application/prescription_service.go`
- `internal/delivery/http/api/prescription_handlers.go`
- `internal/infrastructure/postgres/prescription_projection_repository.go`

---

## Issue C2 (#159) — Backend: endpoint de impresión de receta HTML bajo demanda

**Rama:** `issue/159-prescription-print-endpoint`
**Commit:** `8e01118`
**Merge a main:** `2cef192`
**Capa:** Delivery/HTTP. No toca Core ni Shaders.

### Problema
No existía endpoint para generar el PDF/HTML imprimible de una receta. El archivo `prescription_print.go` existía en el respaldo pero dependía de `consultation_id` y no estaba trackeado en git.

### Solución
Se reincorporó `prescription_print.go` con las siguientes mejoras:
- Usa `PrescriptionService.FindByID` en vez de `ListAll` (más eficiente).
- Valida el token internamente — acepta header `Authorization: Bearer` o query param `?token=` — porque se abre en pestaña nueva con `window.open` donde no se puede adjuntar headers custom.
- Se actualizó `prescriptionAuthDispatcher` en `router.go` para que las peticiones a `/:id/print` no pasen por `JWTMiddleware` general (que solo acepta header), sino directamente al handler que hace su propia validación.
- El PDF no se persiste: se genera bajo demanda en cada solicitud, sin costo de storage en el VPS.

### Archivos involucrados
- `internal/delivery/http/api/prescription_print.go` (nuevo)
- `internal/delivery/http/api/router.go`

---

## Issue C3 (#160) — Frontend: botón "Reimprimir" en vista de detalle de receta

**Rama:** `issue/160-prescription-reprint-button`
**Commit:** `46db613`
**Merge a main:** `d6ca46e`
**Capa:** Asteroide `crm_ui`.

### Solución
- `PrescriptionDetailView.vue` — se agregó botón "Reimprimir" en el header de la sección, que llama a `window.open(/api/v1/prescriptions/:id/print?token=..., '_blank')`.
- `frontend/src/domain/types/prescription.ts` — se agregó `consultation_id?: string` a `Prescription`.

### Archivos involucrados
- `frontend/src/presentation/views/PrescriptionDetailView.vue`
- `frontend/src/domain/types/prescription.ts`

---

## Issue C4 (#161) — Frontend: manejo de errores HTTP en httpClient

**Rama:** `issue/161-httpclient-error-handling`
**Commit:** `16e0ea7`
**Merge a main:** `fcd7397`
**Capa:** Asteroide `crm_ui` / infraestructura frontend.

### Problema
El cliente HTTP devolvía el JSON tal cual aunque `res.ok` fuera false, dejando que el error llegara silencioso o mal formateado al caller.

### Solución
Se reemplazó el return final de `request()` para parsear el JSON primero, y si `!res.ok` lanzar un `Error` con el mensaje del servidor (`json?.error?.message || json?.error?.code || HTTP ${status}`).

### Archivos involucrados
- `frontend/src/infrastructure/api/httpClient.ts`

---

## Issue C5 (#162) — Frontend: nueva consulta con UX mejorado

**Rama:** `issue/162-consultation-new-ux`
**Commit:** `33fdae5`
**Merge a main:** `6e65f11`
**Capa:** Asteroide `crm_ui`.

### Cambios
- T/A separado en dos campos (sistólica/diastólica) para evitar errores de formato.
- Auto-formato de Temperatura (365 → 36.5) y Talla (170 → 1.70) mientras se escribe.
- Solo números en FC, FR, Peso, SAO2.
- Modal de confirmación "¿Guardar sin receta?" antes de proceder si no se adjuntó receta.
- La receta creada dentro de una consulta se vincula automáticamente vía `consultation_id`.
- Si la consulta incluye receta, al guardar se abre el PDF en pestaña nueva automáticamente.
- `PrescriptionRequest` incluye `consultation_id?: string`.

### Archivos involucrados
- `frontend/src/presentation/views/ConsultationNewView.vue`
- `frontend/src/domain/types/prescription.ts`

---

## Issue #163 — Frontend: botón imprimir en todos los contextos

**Rama:** `issue/163-reprint-button-all-contexts`
**Commit:** `2c42b8d`
**Merge a main:** `0886523`
**Capa:** Asteroide `crm_ui`.

### Problema
El botón de imprimir solo existía en la vista de detalle de receta (`/prescriptions/:id`). El médico necesita poder imprimir desde cualquier contexto donde aparezca una receta.

### Solución
Se agregó botón "Imprimir" (SVG lineal de impresora, sin emojis) en tres contextos:
- **Expediente del paciente** (`PatientDetailView.vue`) — en cada card de la sección "Recetas electrónicas".
- **Detalle de consulta** (`ConsultationDetailView.vue`) — carga la receta vinculada vía `prescriptionRepository.listAll()` filtrando por `consultation_id`, y muestra botón si existe.
- **Lista de recetas** (`PrescriptionListView.vue`) — botón con `@click.stop` para no interferir con la navegación al detalle.

### Archivos involucrados
- `frontend/src/presentation/views/PatientDetailView.vue`
- `frontend/src/presentation/views/ConsultationDetailView.vue`
- `frontend/src/presentation/views/PrescriptionListView.vue`


---

## Issue #200 — Modelo de datos de Tenant + Shader Stack base

**Rama:** `issue/200-modelo-datos-tenant`
**Commit:** `c4fd634`
**Merge a main:** `101b9f5`
**Capa:** Infraestructura / Base de datos. No toca Core ni Shaders.
**ADR:** ADR-0025

### Problema
No existía tabla `tenants`. El tenant era un `tenant_id` suelto propagado desde `users` sin integridad referencial. El Shader Stack (`clinical_shader_key`, `export_shader_key`, `extra_shader_keys[]`) no tenía soporte de datos.

### Solución
- Migración `000019_create_tenants.up.sql`: tabla `tenants` con `tenant_id` (PK), `tenant_area`, `country_code`, `clinical_shader_key`, `export_shader_key`, timestamps.
- Tabla `tenant_extra_shaders` para el array 0..N de extra shaders, fail-closed (`active = FALSE` por defecto).
- FK `users.tenant_id → tenants.tenant_id` (aditiva, sobre datos ya poblados).
- Backfill de los 7 tenants existentes con defaults canónicos: `tenant_area='medicine'`, `country_code='MX'`, `clinical_shader_key='med_basic'`, `export_shader_key='export_none'`.

### Archivos involucrados
- `database/migrations/000019_create_tenants.up.sql`

### Fuera de alcance
- Código Go, endpoints, Shaders, UI. Solo esquema de datos + backfill.

---

## Issue #201 — Referencia de vendedor en tenant (provisional fase 1)

**Rama:** `issue/201-referencia-vendedor`
**Commit:** `015a0db`
**Merge a main:** `8b17ddc`
**Capa:** Infraestructura / Base de datos. No toca Core ni Shaders.
**ADR:** ADR-0026

### Problema
No había forma de registrar qué vendedor originó cada tenant para atribución comercial en la fase 1.

### Solución
- Migración `000020_create_vendors.up.sql`: tabla `vendors` con `vendor_id` (formato `vndrNNN`), `name`, `active`, `created_at`.
- Seed inicial: `vndr001 / Carlos Ramírez Herrera`.
- Columna `vendor_ref` (nullable) en `tenants` con FK → `vendors.vendor_id`.

### Archivos involucrados
- `database/migrations/000020_create_vendors.up.sql`

### Fuera de alcance
- Lógica comercial (comisiones, reportes). Core y Shaders no tocan `vendor_ref`.

---

## Issue #202 — Catálogo de shaders de país + MxMedicalShader + delegación NOM-024

**Rama:** `issue/202-shader-catalog-mx-medical`
**Commit:** `931ec27`
**Merge a main:** `46f2abf`
**Capa:** Shaders. No toca Core ni migraciones.
**ADR:** ADR-0002

### Problema
No existía catálogo tipado de shader keys en Go. La validación NOM-024 (cédula + especialidad) vivía incorrectamente en `prescription_handlers.go` (capa `delivery/`), fuera de la capa Shaders.

### Solución
- `internal/shaders/catalog.go`: constantes `ShaderGenericCRM`, `ShaderMxMedical`, `ShaderMxTelemedicine2026` (reservado), y `ShaderRegistry` que resuelve el shader correcto por key. Fail-closed.
- `internal/shaders/mx_medical.go`: `MxMedicalShader` + `ValidateMxMedicalProfile` con validación NOM-024-SSA3-2012.
- `prescription_handlers.go`: validación NOM-024 inline reemplazada por delegación a `shaders.ValidateMxMedicalProfile`.

### Archivos involucrados
- `internal/shaders/catalog.go` (nuevo)
- `internal/shaders/mx_medical.go` (nuevo)
- `internal/delivery/http/api/prescription_handlers.go`

### Fuera de alcance
- Core, migraciones, endpoints, UI. `mx_telemedicine_2026` declarado pero sin implementación activa.

---

## Issue #203 — Delegación NOM-024 en admin handler

**Rama:** `issue/203-nom024-admin-handler`
**Commit:** `886eeed`
**Merge a main:** `2c7a278`
**Capa:** Delivery / API. No toca Core ni Shaders.
**ADR:** ADR-0002

### Problema
`admin_handlers.go` contenía validación NOM-024 (cédula + especialidad) inline en el `switch` de `HandleAdminCreateUser`, fuera de la capa Shaders.

### Solución
- Eliminadas las dos líneas NOM-024 del `switch`.
- Agregada llamada a `shaders.ValidateMxMedicalProfile` después del `switch`, antes de `ExistsByEmail`.
- Agregado import de `shaders` al archivo.

### Archivos involucrados
- `internal/delivery/http/api/admin_handlers.go`

### Fuera de alcance
- Core, migraciones, Shaders, UI. Sin cambio de firma de endpoints.

---

## Issue #204 — TenantRepository + ShaderService dinámico por clinical_shader_key

**Rama:** `issue/204-shader-stack-tenant`
**Commit:** `9119524`
**Merge a main:** `ef497f9`
**Capa:** Aplicación / Infraestructura / Delivery. No toca Core ni migraciones.
**ADR:** ADR-0025 + ADR-0002

### Problema
`ShaderService` hardcodeaba siempre `NewMedicalBasicShader()` sin consultar el `clinical_shader_key` real del tenant. No existía puerto ni adaptador para leer la tabla `tenants`.

### Solución
- `ports/tenant_repository.go`: interfaz `TenantRepository` con `GetByID`.
- `postgres/tenant_repository.go`: adaptador que lee `tenants` por `tenant_id`. Fail-closed.
- `delivery_deps.go`: `DeliveryDeps` con `TenantRepo` para el paquete `delivery`.
- `shader_service.go`: `Authorize()` resuelve `clinical_shader_key` del tenant via `ShaderRegistry`. Fail-closed: tenant no encontrado → `DecisionDeny`.
- Handlers ECE (`ece_handlers.go`, `ece_issue_handlers.go`, `ece_void_handlers.go`, `ece_export_handlers.go`, `ece_draft_save_handler.go`): `NewShaderService(deliveryDeps.TenantRepo)`.
- `main.go`: `InitDeliveryDeps` + `TenantRepo` inyectado en `api.Deps`.

### Archivos involucrados
- `internal/application/ports/tenant_repository.go` (nuevo)
- `internal/infrastructure/postgres/tenant_repository.go` (nuevo)
- `internal/delivery/http/delivery_deps.go` (nuevo)
- `internal/delivery/http/shader_service.go`
- `internal/delivery/http/ece_handlers.go`
- `internal/delivery/http/ece_issue_handlers.go`
- `internal/delivery/http/ece_void_handlers.go`
- `internal/delivery/http/ece_export_handlers.go`
- `internal/delivery/http/ece_draft_save_handler.go`
- `cmd/vuhmik-api/main.go`

### Fuera de alcance
- Core, migraciones, UI. `export_shader_key` y `extra_shaders` conectados en issues siguientes.

---

## Issue #205 — ExportShaderRegistry + resolución dinámica de export shader

**Rama:** `issue/205-export-shader-catalog`
**Commit:** `09f0a6c`
**Merge a main:** `fbb653e`
**Capa:** Shaders / Delivery / API. No toca Core ni migraciones.
**ADR:** ADR-0002

### Problema
`ShaderService.Export()` y `buildExportShader()` hardcodeaban `NewLegalExportShader()` sin consultar `export_shader_key` del tenant. No existía catálogo tipado de export shader keys.

### Solución
- `catalog.go`: constantes `ExportShaderLegal` y `ExportShaderNone`, y `ExportShaderRegistry` que resuelve el export shader por key. Fail-closed: key desconocido → nil (el caller deniega).
- `shader_service.go`: `Export()` resuelve `export_shader_key` del tenant dinámicamente.
- `api/deps.go`: agregado `TenantRepo ports.TenantRepository`.
- `evidence_handlers.go`: agregada `buildExportShaderForTenant(tenantID)` con resolución dinámica.
- `main.go`: `TenantRepo` inyectado en `api.Deps`.

### Archivos involucrados
- `internal/shaders/catalog.go`
- `internal/delivery/http/shader_service.go`
- `internal/delivery/http/api/deps.go`
- `internal/delivery/http/api/evidence_handlers.go`
- `cmd/vuhmik-api/main.go`

### Fuera de alcance
- Core, migraciones, UI. `extra_shaders` conectado en issue siguiente.

---

## Issue #206 — Extra shaders dinámicos por tenant + evaluación encadenada

**Rama:** `issue/206-extra-shaders-tenant`
**Commit:** `2dfc842`
**Merge a main:** `6e0b70b`
**Capa:** Aplicación / Infraestructura / Delivery. No toca Core ni migraciones.
**ADR:** ADR-0025

### Problema
`tenant_extra_shaders` existía en la base pero ningún código Go la consultaba. El Shader Stack estaba incompleto: los extra shaders (0..N) no se evaluaban en ninguna operación.

### Solución
- `ports/tenant_repository.go`: agregado `ExtraShaderKeys []string` a `TenantConfig`.
- `postgres/tenant_repository.go`: segunda query en `GetByID` que lee `tenant_extra_shaders` donde `active = true`.
- `shader_service.go`: `Authorize()` evalúa extra shaders encadenados después del clinical shader. Fail-closed: cualquier extra shader que deniega detiene la cadena.

### Archivos involucrados
- `internal/application/ports/tenant_repository.go`
- `internal/infrastructure/postgres/tenant_repository.go`
- `internal/delivery/http/shader_service.go`

### Fuera de alcance
- Core, UI, endpoints. Sin nuevas migraciones.

---

## Issue #207 — Activar mx_medical y legal_export en tenants existentes

**Rama:** `issue/207-activar-mx-medical-tenants`
**Commit:** `5cb706c`
**Merge a main:** `bfb8202`
**Capa:** Base de datos. No toca código Go ni Shaders.
**ADR:** ADR-0025 + ADR-0002

### Problema
Los 7 tenants existentes tenían `export_shader_key='export_none'` y `tenant_extra_shaders` vacía. El cumplimiento NOM-024 estaba implementado pero no activado.

### Solución
- Migración `000021_seed_mx_medical_tenants.up.sql`:
  - INSERT en `tenant_extra_shaders`: `mx_medical / active=TRUE` para los 7 tenants.
  - UPDATE en `tenants`: `export_shader_key='legal_export'` donde era `'export_none'`.

### Archivos involucrados
- `database/migrations/000021_seed_mx_medical_tenants.up.sql`

### Fuera de alcance
- Código Go, Shaders, Core, UI.

---

## Issue #208 — Conectar buildExportShaderForTenant al handler de export

**Rama:** `issue/208-connect-export-shader-tenant`
**Commit:** `da58dc8`
**Merge a main:** `3c63aa5`
**Capa:** Delivery / API. No toca Core ni Shaders.
**ADR:** ADR-0002

### Problema
`HandleEvidenceExport` llamaba a `buildExportShader()` (fallback estático) en vez de `buildExportShaderForTenant(tenantID)` (resolución dinámica creada en #205).

### Solución
- Una línea en `evidence_handlers.go`: `buildExportShader()` → `buildExportShaderForTenant(tenantID)`.

### Archivos involucrados
- `internal/delivery/http/api/evidence_handlers.go`

### Fuera de alcance
- Todo lo demás. Cambio quirúrgico de una línea.

---

## Issue #209 — Validación explícita de clinical_shader_key contra catálogo

**Rama:** `issue/209-validar-clinical-shader-key`
**Commit:** `784be12`
**Merge a main:** `d5c3042`
**Capa:** Shaders / Delivery. No toca Core ni migraciones.
**ADR:** ADR-0002

### Problema
`ShaderRegistry.Resolve()` hacía fallback silencioso a `med_basic` para cualquier key desconocido, sin reportar la anomalía. Un `clinical_shader_key` corrupto en la base pasaba desapercibido.

### Solución
- `catalog.go`: `KnownShaderKeys` (mapa de keys activos válidos) e `IsKnownShaderKey()`.
- `shader_service.go`: `Authorize()` verifica `IsKnownShaderKey()` antes de resolver. Key inválido → `DecisionDeny` con `ER-SHADER-002`.

### Archivos involucrados
- `internal/shaders/catalog.go`
- `internal/delivery/http/shader_service.go`

### Fuera de alcance
- Core, migraciones, UI, endpoints.

---

## Issue #212 — IPS Bundle FHIR R4 como formato canónico de export

**Rama:** `issue/212-ips-fhir-export`
**Commit:** `8dbe78a`
**Merge a main:** `473f4fe`
**Capa:** Shaders. No toca Core ni migraciones.
**ADR:** ADR-0010

### Problema
El export JSON producía un dump plano de ExportData. El export XML usaba el esquema CDA propio (ADR-0007 legado). Ninguno era IPS/FHIR R4.

### Solución
- `internal/shaders/ips_bundle.go` (nuevo): `IPSBundle`, `BuildIPSBundle`, `MarshalIPSBundleJSON`, `MarshalIPSBundleXML`.
- `legal_export.go`: `GenerateExport` ahora produce IPS Bundle JSON.
- `legal_export_xml.go`: `GenerateExportXML` ahora produce IPS Bundle XML. El esquema CDA propio queda como `generateExportXMLLegacy` deprecado.

### Archivos involucrados
- `internal/shaders/ips_bundle.go` (nuevo)
- `internal/shaders/legal_export.go`
- `internal/shaders/legal_export_xml.go`

### Fuera de alcance
- Core, migraciones, UI, endpoints. Las secciones IPS de módulos específicos se agregan en issues posteriores.

---

## Issue #213 — Audit Package ZIP síncrono por paciente

**Rama:** `issue/213-audit-package-zip`
**Commit:** `981a4e3`
**Merge a main:** `a9ed7fe`
**Capa:** Shaders / Delivery / API. No toca Core.
**ADR:** ADR-0027 (nuevo)

### Problema
No existía endpoint para exportar el expediente completo de un paciente como paquete auditable. No había ADR que autorizara ZIP síncrono en v1.

### Solución
- `docs/adr/ADR-0027-audit-package-zip-sync.md`: ZIP síncrono en v1, en memoria, sin persistencia. Límite: un paciente a la vez.
- `internal/shaders/audit_zip.go` (nuevo): `BuildAuditPackageZIP`, estructura `manifest.json + ips/ + evidence/ + hashes/`.
- `patient_handlers.go`: `HandlePatientExportZIP` — `GET /api/v1/patients/:id/export/zip`.
- `router.go`: `case "export/zip"`.

### Archivos involucrados
- `docs/adr/ADR-0027-audit-package-zip-sync.md` (nuevo)
- `internal/shaders/audit_zip.go` (nuevo)
- `internal/delivery/http/api/patient_handlers.go`
- `internal/delivery/http/api/router.go`

### Fuera de alcance
- Sistema de jobs asíncronos (requiere ADR posterior). Core, UI.

---

## Issue #214 — Módulo de diagnósticos CIE-10 + IPS Condition

**Rama:** `issue/214-diagnosticos-cie10`
**Commit:** `acf3ffa`
**Merge a main:** `807f14f`
**Capa:** Shaders / API. Migración en issue #215.
**ADR:** ADR-0013

### Problema
El módulo de diagnósticos estructurados (ADR-0013) no estaba implementado. No existía shader, handler ni proyector IPS para diagnósticos CIE-10.

### Solución
- `diagnosis_shader.go`: `DiagnosisContent`, `ValidateDiagnosisContent`, `BuildDiagnosisBlob`.
- `ips_diagnosis_export.go`: `IPSCondition`, `ProjectDiagnosisToIPS`, `ExportDiagnosisAsIPS`.
- `diagnosis_handlers.go`: `HandleDiagnosisCreate` y `HandleDiagnosisListByPatient`.
- `router.go`: `case "diagnoses"`.

### Archivos involucrados
- `internal/shaders/diagnosis_shader.go` (nuevo)
- `internal/shaders/ips_diagnosis_export.go` (nuevo)
- `internal/delivery/http/api/diagnosis_handlers.go` (nuevo)
- `internal/delivery/http/api/router.go`

---

## Issue #215 — Fix migración 000022 proyección diagnósticos

**Commit:** `dc56138`
**Capa:** Base de datos.
**ADR:** ADR-0013

### Problema
La migración `000022_diagnosis_projections.up.sql` no se commiteó en el Issue #214.

### Solución
- Creación y aplicación de `000022_diagnosis_projections.up.sql`: tabla `diagnosis_projections` con índices.

### Archivos involucrados
- `database/migrations/000022_diagnosis_projections.up.sql` (nuevo)

---

## Issue #216 — Módulo de inmunizaciones + IPS Immunization

**Rama:** `issue/216-inmunizaciones`
**Commit:** `6e38988`
**Merge a main:** `122b32e`
**Capa:** Shaders / API / BD. No toca Core.
**ADR:** ADR-0014

### Problema
El módulo de inmunizaciones (ADR-0014) no estaba implementado.

### Solución
- `immunization_shader.go`: `ImmunizationContent`, `ValidateImmunizationContent`, `BuildImmunizationBlob`.
- `ips_immunization_export.go`: `IPSImmunization`, `IPSAnnotation`, `ProjectImmunizationToIPS`.
- `immunization_handlers.go`: `HandleImmunizationCreate` y `HandleImmunizationListByPatient`.
- `000023_immunization_projections.up.sql`: tabla `immunization_projections`.
- `router.go`: `case "immunizations"`.

### Archivos involucrados
- `database/migrations/000023_immunization_projections.up.sql` (nuevo)
- `internal/delivery/http/api/immunization_handlers.go` (nuevo)
- `internal/delivery/http/api/router.go`
- `internal/shaders/immunization_shader.go` (nuevo)
- `internal/shaders/ips_immunization_export.go` (nuevo)

---

## Issue #217 — Módulo de resultados de laboratorio + IPS Observation

**Rama:** `issue/217-laboratorio`
**Commit:** `672dbc4`
**Merge a main:** `ce2dd15`
**Capa:** Shaders / API / BD. No toca Core.
**ADR:** ADR-0015

### Problema
El módulo de resultados de laboratorio (ADR-0015) no estaba implementado.

### Solución
- `lab_result_shader.go`: `LabResultContent`, `ValidateLabResultContent`, `BuildLabResultBlob`.
- `ips_lab_result_export.go`: `IPSObservation`, `ProjectLabResultToIPS`, `ExportLabResultAsIPS`.
- `lab_result_handlers.go`: `HandleLabResultCreate` y `HandleLabResultListByPatient`.
- `000024_lab_result_projections.up.sql`: tabla `lab_result_projections`.
- `router.go`: `case "lab-results"`.

### Archivos involucrados
- `database/migrations/000024_lab_result_projections.up.sql` (nuevo)
- `internal/delivery/http/api/lab_result_handlers.go` (nuevo)
- `internal/delivery/http/api/router.go`
- `internal/shaders/ips_lab_result_export.go` (nuevo)
- `internal/shaders/lab_result_shader.go` (nuevo)

---

## Issue #219 — Proyector IPS MedicationStatement para recetas

**Commit:** `783108f`
**Merge a main:** `5c6900a`
**Capa:** Shaders. No toca Core.
**ADR:** ADR-0011

### Problema
El módulo de recetas (`prescription_projections`, ya en producción desde antes) no tenía proyector a IPS/FHIR R4. El export legal de un paciente con recetas emitidas no incluía `MedicationStatement`.

### Solución
- `ips_prescription_export.go` (nuevo): proyector de `prescription_projections` a `MedicationStatement` conforme perfil IPS.

### Archivos involucrados
- `internal/shaders/ips_prescription_export.go` (nuevo)

---

## Issue #220 — vendor_ref: handler y catálogo de vendedores

**Commit:** `19baa41`
**Merge a main:** `7d492b7`
**Capa:** Aplicación / Infraestructura / API.
**ADR:** ADR-0026

### Problema
La columna `vendor_ref` (migración `000020_create_vendors.up.sql`, ya aplicada) no tenía forma de asignarse ni consultarse desde la API administrativa.

### Solución
- `vendor_repository.go` (puerto + adaptador Postgres, nuevos): catálogo de vendedores.
- `admin_handlers.go`: endpoint para asignar `vendor_ref` a un tenant.
- `vendor_handlers.go` (nuevo): `SetVendorRef`.
- `tenant_repository.go` (puerto + adaptador): soporte de lectura/escritura de `vendor_ref`.

### Archivos involucrados
- `cmd/vuhmik-api/main.go`
- `internal/application/ports/tenant_repository.go`
- `internal/application/ports/vendor_repository.go` (nuevo)
- `internal/delivery/http/api/admin_handlers.go`
- `internal/delivery/http/api/deps.go`
- `internal/delivery/http/api/router.go`
- `internal/delivery/http/api/vendor_handlers.go` (nuevo)
- `internal/infrastructure/postgres/tenant_repository.go`
- `internal/infrastructure/postgres/vendor_repository.go` (nuevo)

---

## Issue #221 — Traspaso de paciente: import + FindByCURP

**Commit:** `b2f5f0b`
**Merge a main:** `6db2446`
**Capa:** Delivery / Infraestructura. No toca Core.
**ADR:** ADR-0009

### Problema
ADR-0009 (traspaso de paciente entre tenants) no tenía implementación. No existía endpoint de import ni búsqueda de paciente por CURP en el tenant destino.

### Solución
- `patient_import_handler.go` (nuevo): `POST /api/v1/patients/import` — creación de evidencia en estado `issued` a partir de un paquete de traspaso.
- `patient_repository.go`: se agregó `FindByCURP` para localizar al paciente destino antes de importar.

### Archivos involucrados
- `internal/delivery/http/api/patient_import_handler.go` (nuevo)
- `internal/delivery/http/api/router.go`
- `internal/infrastructure/postgres/patient_repository.go`

---

## Issue #222 — Export de paquete de traspaso compatible con import

**Commit:** `9714142`
**Merge a main:** `2773d1e`
**Capa:** Delivery. No toca Core.
**ADR:** ADR-0009

### Problema
El endpoint de import (issue #221) no tenía contraparte de export: no existía forma de generar, desde el tenant origen, el paquete de traspaso que el import del issue #221 pudiera consumir.

### Solución
- `patient_transfer_export_handler.go` (nuevo): genera el paquete de traspaso en el formato esperado por `patient_import_handler.go`.

### Archivos involucrados
- `internal/delivery/http/api/patient_transfer_export_handler.go` (nuevo)
- `internal/delivery/http/api/router.go`

---

## Issue #223 — Importación de IPS Bundle FHIR R4 externo (IMSS/ISSSTE)

**Commit:** `a5a6d80`
**Merge a main:** `cff2fd5`
**Capa:** Delivery. No toca Core.
**ADR:** ADR-0028 (nuevo)

### Problema
El import de paciente (issue #221) solo aceptaba el formato propio `vuhmik-transfer-v1`. No había forma de recibir un expediente externo en formato IPS/FHIR R4 estándar (IMSS, ISSSTE u otro emisor).

### Solución
- `docs/adr/ADR-0028-ips-fhir-external-import.md` (nuevo): acepta cualquier IPS Bundle FHIR R4; recursos no reconocidos se preservan como blob `fhir_unknown` con su contenido íntegro; el origen se marca (`fhir-imss`, `fhir-issste`, `fhir-external`); requiere CURP pre-registrado.
- `ips_external_parser.go` (nuevo): parser del Bundle FHIR R4 externo.
- `patient_import_handler.go`: detección automática de formato por `resourceType: "Bundle"` vs. `format: "vuhmik-transfer-v1"`.

### Archivos involucrados
- `docs/adr/ADR-0028-ips-fhir-external-import.md` (nuevo)
- `internal/delivery/http/api/ips_external_parser.go` (nuevo)
- `internal/delivery/http/api/patient_import_handler.go`

---

## Issue #224 — ADR-0012 → Aceptado (alergias e intolerancias)

**Commit:** `01bc58e`
**Capa:** Documentación. Sin cambios de código.
**ADR:** ADR-0012

### Problema
ADR-0012 documentaba el módulo de alergias como decisión, pero no reflejaba que la implementación (migración `000011_projections`, `allergy_shader.go`, `ips_allergy_export.go`, `allergy_handlers.go`, `allergy_service.go`) ya estaba en `main` desde los issues #132, #133 y #135.

### Solución
- `ADR-0012-alergias-intolerancias.md`: estado actualizado a Aceptado, con sección "Estado de implementación" documentando los archivos reales.

### Archivos involucrados
- `docs/adr/ADR-0012-alergias-intolerancias.md`

---

## Issue #226 — Worker de métricas (WAR-A)

**Commit:** `bc3fdd4`
**Capa:** Workers. No toca Core ni Shaders.
**ADR:** ADR-0019

### Problema
ADR-0019 (panel de métricas de negocio) no tenía worker que calculara y persistiera snapshots. La tabla `metrics_snapshot` (migración `000013`, ya existente) no se poblaba.

### Solución
- `metrics_worker.go` (nuevo): `MetricsWorker` calcula snapshot cada 4 horas (cuentas totales/activas/suspendidas, MRR, pacientes, notas, alergias, recetas, distribución de módulos) y lo persiste; `MetricsPurgeWorker` elimina snapshots con más de 30 días.
- `main.go`: registro de ambos workers con apagado ordenado ante señales del sistema operativo (WAR-A).

### Archivos involucrados
- `cmd/vuhmik-api/main.go`
- `internal/workers/metrics_worker.go` (nuevo)

---

## Issue #227 — Handlers GET de métricas de administración

**Commit:** `219a577`
**Capa:** Delivery / API. No toca Core.
**ADR:** ADR-0019

### Problema
El snapshot de métricas (issue #226) no tenía forma de consultarse desde la API. No existían endpoints de solo lectura para el panel de administración.

### Solución
- `metrics_handlers.go` (nuevo): `HandleAdminMetrics` (resumen agregado), `HandleAdminMetricsAccounts` (lista de cuentas con conteos), `HandleAdminMetricsAccountDetail` (detalle de una cuenta), `HandleAdminMetricsModules` (distribución de módulos).
- `deps.go`: se agregó `deps.DB *pgxpool.Pool` para que los handlers de métricas consulten directamente el snapshot.
- `router.go`: rutas `/api/v1/admin/metrics*` protegidas por `AdminMiddleware`.

### Archivos involucrados
- `cmd/vuhmik-api/main.go`
- `internal/delivery/http/api/deps.go`
- `internal/delivery/http/api/metrics_handlers.go` (nuevo)
- `internal/delivery/http/api/router.go`

---

## Issue #228 — ADR-0019 → Aceptado (panel de métricas)

**Commit:** `06e4bda`
**Capa:** Documentación. Sin cambios de código.
**ADR:** ADR-0019

### Problema
ADR-0019 no reflejaba la implementación real completada en los issues #226 y #227.

### Solución
- `ADR-0019-panel-metricas.md`: estado actualizado a Aceptado, con sección "Estado de implementación" documentando migración, workers y handlers reales.

### Archivos involucrados
- `docs/adr/ADR-0019-panel-metricas.md`

---

## Issue #229 — Migración activity_snapshot (000026)

**Commit:** `68f9eb6`
**Capa:** Base de datos.
**ADR:** ADR-0023 (nuevo)

### Problema
ADR-0023 (panel de actividad y uso) requería una tabla de snapshot por tenant y periodo, distinta de `metrics_snapshot` (ADR-0019, que es global). No existía tabla para este propósito.

### Solución
- `000026_activity_snapshot.up.sql` (nuevo): tabla `activity_snapshot` (PK compuesta `tenant_id, period`) con conteos de notas, alergias, recetas, exportaciones, pacientes y sesiones por mes, más índices por tenant/periodo.

### Archivos involucrados
- `database/migrations/000026_activity_snapshot.up.sql` (nuevo)

---

## Issue #230 — activity_log helper + eventos de sesión

**Commit:** `6c582fa`
**Capa:** Delivery. No toca Core.
**ADR:** ADR-0023

### Problema
No existía forma de registrar eventos de actividad (inicio/fin de sesión) sin PHI para alimentar `activity_snapshot` (issue #229).

### Solución
- `activity_log.go` (nuevo): helper `logActivity(ctx, tenantID, eventType)` — inserta en `activity_log` solo `tenant_id`, tipo de evento y timestamp; un fallo de escritura no bloquea el flujo principal.
- `auth_handlers.go`: se agregaron eventos `session_start` en `HandleLogin` y `session_end` en `HandleLogout`.

### Archivos involucrados
- `internal/delivery/http/api/activity_log.go` (nuevo)
- `internal/delivery/http/api/auth_handlers.go`

---

## Issue #231 — Handlers GET de actividad de administración

**Commit:** `ec710ec`
**Capa:** Delivery / API. No toca Core.
**ADR:** ADR-0023

### Problema
Los eventos registrados (issue #230) y el snapshot (issue #229) no tenían forma de consultarse desde la API administrativa.

### Solución
- `activity_handlers.go` (nuevo): `HandleAdminActivity` (`GET /api/v1/admin/activity` — lista de tenants con conteos agregados desde `activity_snapshot`), `HandleAdminActivityDetail` (`GET /api/v1/admin/activity/:tenant` — detalle por mes, últimos 12 meses).
- `router.go`: rutas protegidas por `AdminMiddleware`.

### Archivos involucrados
- `internal/delivery/http/api/activity_handlers.go` (nuevo)
- `internal/delivery/http/api/router.go`

---

## Issue #232 — ADR-0023 → Aceptado (panel de actividad)

**Commit:** `a9fc20b`
**Capa:** Documentación. Sin cambios de código.
**ADR:** ADR-0023

### Problema
ADR-0023 no reflejaba la implementación real completada en los issues #229, #230 y #231.

### Solución
- `ADR-0023-panel-actividad.md`: estado actualizado a Aceptado, con sección "Estado de implementación" documentando migraciones, helper de log y handlers reales.

### Archivos involucrados
- `docs/adr/ADR-0023-panel-actividad.md`

---

## Issue #233 — Panel admin: métricas + actividad + navegación lateral

**Commit:** `d682f70`
**Capa:** Asteroide (frontend admin). No toca Core ni Shaders.
**ADR:** ADR-0019, ADR-0023

### Problema
Los endpoints de métricas (#227) y actividad (#231) no tenían consumo en el frontend. El panel de administración no mostraba esta información ni tenía navegación hacia ella.

### Solución
- `AdminView.vue`: se agregaron las secciones de Métricas y Actividad al panel de administración, con navegación lateral hacia ambas.

### Archivos involucrados
- `frontend/src/presentation/views/AdminView.vue`

---

## Issue #234 — ADR-0009, 0011, 0017, 0018, 0021 → Aceptado (encabezados desactualizados)

**Commit:** `bb34d10`
**Merge a main:** `b953104`
**Capa:** Documentación. Sin cambios de código.
**ADR:** ADR-0009, ADR-0011, ADR-0017, ADR-0018, ADR-0021

### Problema
Auditoría completa de los 28 ADRs contra código real reveló que 5 ADRs tenían
código ya funcional en `main` pero sus encabezados seguían diciendo "Propuesto" /
"No implementado". La auditoría también encontró 3 gaps reales documentados como
errores de cobertura (ver issues #235, #236).

### Solución
- ADR-0009, 0011, 0017, 0018, 0021: encabezado actualizado a Aceptado con sección
  "Estado de implementación" documentando archivos e issues reales.
- ADR-0011/0021: gap falso de verificación cruzada perfil-receta corregido — la
  verificación ya existía vía `mx_medical.go` (issue #202), error de auditoría inicial.
- ADR-0017: cobertura de CapabilityGuard documentada como parcial (diagnosis,
  immunization, lab_result sin compuerta) — resuelto en issue #235.

### Archivos involucrados
- `docs/adr/ADR-0009-patient-transfer.md`
- `docs/adr/ADR-0011-medicacion-receta.md`
- `docs/adr/ADR-0017-registro-capacidades.md`
- `docs/adr/ADR-0018-panel-toggles.md`
- `docs/adr/ADR-0021-perfil-profesional.md`

---

## Issue #235 — CapabilityGuard + Service + proyecciones CQRS + Void para diagnosis, immunization, lab_result

**Commit:** `7792c38`
**Merge a main:** `309b2e2`
**Capa:** Aplicación / Infraestructura / API. No toca Core.
**ADR:** ADR-0013, ADR-0014, ADR-0015, ADR-0017, ADR-0022

### Problema
Los módulos diagnosis, immunization y lab_result tenían Shaders definidos pero
completamente desconectados. Los handlers iban directo a `EvidenceRepo` sin pasar
por ningún Shader ni CapabilityGuard. Las tablas de proyección CQRS (migraciones
000022/023/024) existían pero ningún código las escribía ni leía. No existía Void
para ninguno de los tres módulos.

### Solución
- 3 puertos de proyección nuevos (`DiagnosisProjectionRepository`,
  `ImmunizationProjectionRepository`, `LabResultProjectionRepository`).
- 3 adaptadores Postgres para las tablas de proyección ya existentes.
- 3 servicios (`DiagnosisService`, `ImmunizationService`, `LabResultService`)
  siguiendo el patrón de `AllergyService`: Create con CapabilityGuard,
  ListByPatient leyendo de la proyección, Void.
- Handlers reescritos para delegar al Service.
- Rutas `/void` nuevas para los 3 módulos.
- `deps.go` y `main.go` actualizados.

### Archivos involucrados (15 archivos)
- `internal/application/diagnosis_service.go` (nuevo)
- `internal/application/immunization_service.go` (nuevo)
- `internal/application/lab_result_service.go` (nuevo)
- `internal/application/ports/diagnosis_projection_repository.go` (nuevo)
- `internal/application/ports/immunization_projection_repository.go` (nuevo)
- `internal/application/ports/lab_result_projection_repository.go` (nuevo)
- `internal/infrastructure/postgres/diagnosis_projection_repository.go` (nuevo)
- `internal/infrastructure/postgres/immunization_projection_repository.go` (nuevo)
- `internal/infrastructure/postgres/lab_result_projection_repository.go` (nuevo)
- `internal/delivery/http/api/diagnosis_handlers.go`
- `internal/delivery/http/api/immunization_handlers.go`
- `internal/delivery/http/api/lab_result_handlers.go`
- `internal/delivery/http/api/deps.go`
- `internal/delivery/http/api/router.go`
- `cmd/vuhmik-api/main.go`

---

## Issue #236 — CapabilityGuard para note (Create/Edit/Void)

**Commit:** `a65cfdf`
**Merge a main:** `da46976`
**Capa:** Aplicación / Delivery. No toca Core.
**ADR:** ADR-0017, ADR-0022

### Problema
El módulo `note` (notas clínicas) tampoco pasaba por CapabilityGuard — mismo
problema que los 3 módulos del issue #235 pero en un módulo distinto. Los handlers
`HandleEvidenceDraft` y `HandleEvidenceEdit` iban directo a `EvidenceRepo` sin
ninguna compuerta. `HandleEvidenceVoid` (endpoint genérico para cualquier tipo de
evidencia) tampoco tenía guard, y asumía implícitamente que solo se usaba para notas.

### Solución
- `NoteService` nuevo: `Create` (con auto-emit, ADR-0006) y `Edit` (void+replace
  silencioso) ambos con CapabilityGuard usando `MedicalBasicShader` como base.
- `HandleEvidenceDraft` y `HandleEvidenceEdit` delegan en `NoteService`.
- `moduleShaderForBlob()` nuevo: helper que lee el campo `type` real del blob antes
  de aplicar el guard en `HandleEvidenceVoid` — en vez de asumir un módulo fijo,
  resuelve dinámicamente el Shader correcto para cualquier tipo de evidencia.
- `deps.go` y `main.go` actualizados.

### Archivos involucrados (4 archivos)
- `internal/application/note_service.go` (nuevo)
- `internal/delivery/http/api/evidence_handlers.go`
- `internal/delivery/http/api/deps.go`
- `cmd/vuhmik-api/main.go`

---

## Issue #237 — UI para diagnósticos, inmunizaciones y resultados de laboratorio

**Commit:** `4737be1`
**Merge a main:** `91d50db`
**Capa:** Asteroide (frontend). No toca Core ni Shaders.
**ADR:** ADR-0013, ADR-0014, ADR-0015

### Problema
Los 3 módulos del issue #235 tenían backend completo pero sin ninguna vista Vue.
No eran validables desde el navegador.

### Solución
- 3 tipos TypeScript nuevos: `Diagnosis`, `Immunization`, `LabResult`.
- 3 repositorios TypeScript nuevos: `diagnosisRepository`, `immunizationRepository`,
  `labResultRepository` — mismo patrón que `allergyRepository`.
- 3 secciones colapsables en `PatientDetailView.vue` (mismo patrón que alergias):
  formulario inline, create/void, contador visible.
- Validado end-to-end en navegador el 2026-07-09.

### Archivos involucrados (7 archivos)
- `frontend/src/domain/types/diagnosis.ts` (nuevo)
- `frontend/src/domain/types/immunization.ts` (nuevo)
- `frontend/src/domain/types/lab_result.ts` (nuevo)
- `frontend/src/infrastructure/repositories/diagnosisRepository.ts` (nuevo)
- `frontend/src/infrastructure/repositories/immunizationRepository.ts` (nuevo)
- `frontend/src/infrastructure/repositories/labResultRepository.ts` (nuevo)
- `frontend/src/presentation/views/PatientDetailView.vue`

---

## Fixes de sesión 2026-07-09 (bugs detectados durante validación en navegador)

**Commits:** `1713e7c`, `a10006b`, `6a8f43f`, `66b2387`, `efcacc4`
**Capa:** Workers / Delivery / Asteroide.

### Problema
Durante la sesión de validación end-to-end del 2026-07-09 se detectaron 4 bugs
preexistentes que bloqueaban la validación del panel admin:

1. `metrics_worker.go` usaba nombres de columna incorrectos (`monthly_cost`,
   `is_active`, `capability_key`) que no existen en el esquema real de
   `tenant_capabilities` (columnas reales: `costo`, `active`, `module_id`).
2. `metrics_handlers.go` escaneaba `calculated_at` (timestamptz) en un `string` —
   incompatible con pgx v5 que requiere `time.Time` para ese tipo OID.
3. `metrics_handlers.go` escaneaba `mrr` (numeric(10,2)) en `float64` sin castear —
   pgx v5 en modo binario no convierte `numeric` a `float64` automáticamente.
4. `AdminView.vue` tenía dos componentes Vue completos concatenados (dos bloques
   `<script setup>` + `<template>` + `<style>`) — resultado del issue #233 que
   pegó un componente nuevo sobre el existente en vez de fusionarlos. Bloqueaba
   el arranque del frontend completamente.
5. Migración `000026` (`activity_snapshot`) no estaba aplicada en BD local.

### Solución
- `metrics_worker.go`: corregir los 3 nombres de columna incorrectos.
- `metrics_handlers.go`: cambiar `CalculatedAt string` → `time.Time` y agregar
  `mrr::float8` en el query SQL para forzar el cast en origen.
- `AdminView.vue`: eliminar el primer bloque duplicado (332 líneas), dejando solo
  el segundo bloque que era el completo y actualizado.
- BD local: `migrate up` para aplicar `000026_activity_snapshot`.

### Archivos involucrados
- `internal/workers/metrics_worker.go`
- `internal/delivery/http/api/metrics_handlers.go`
- `frontend/src/presentation/views/AdminView.vue`

---

## Issue #238 — Admin: editar perfil, billing mode, reset contraseña y recálculo inmediato de métricas

**Rama:** `issue/238-admin-edit-profile-billing`
**Merge a main:** `efcacc4`
**Capa:** Delivery / Infraestructura / Asteroide (frontend). Sin cambios de Core.

### Problema
El admin no podía editar el perfil profesional de un médico existente, cambiar su
plan de facturación, ni resetear su contraseña sin acceso directo a la BD.
El MRR siempre mostraba $0 porque nadie tenía un plan asignado y el snapshot
solo recalculaba cada 4 horas.

### Solución
- Migración `000027`: columnas `billing_mode` y `monthly_fee` en `users`.
- `UserRepository`: métodos `SetBilling()`, `SetPassword()`, `FindAll()` actualizado.
- 3 handlers nuevos: `HandleAdminUpdateProfile`, `HandleAdminSetBilling`, `HandleAdminResetPassword`.
- `adminUserDispatcher`: enruta `PUT /admin/users/:tenant_id/{profile|billing|password}`.
- `HandleAdminMetricsRecalculate`: `POST /admin/metrics/recalculate` fuerza snapshot inmediato.
- `MetricsWorker`: MRR calcula según `billing_mode` (mensual fijo o suma de módulos).
- Frontend: panel de edición inline en detalle de tenant con 3 botones.

### Archivos involucrados
- `database/migrations/000027_billing_mode.up.sql` (nuevo)
- `internal/infrastructure/postgres/user_repository.go`
- `internal/delivery/http/api/admin_handlers.go`
- `internal/delivery/http/api/metrics_handlers.go`
- `internal/delivery/http/api/router.go`
- `cmd/vuhmik-api/main.go`
- `frontend/src/presentation/views/AdminView.vue`

---

## Issues #239/#240 — Panel Salud de cuentas y Estado del sistema

**Rama:** `issue/239-240-salud-cuentas-sistema`
**Merge a main:** `b64c572`
**Capa:** Workers / Delivery / Asteroide (frontend).

### Problema
El admin no tenía visibilidad de qué médicos estaban activos, en riesgo de churn,
o inactivos, ni del estado operativo del sistema (BD, backups, disco, workers).

### Solución
- Migración `000028`: tablas `account_health_snapshot`, `system_snapshot`, `login_attempts`.
- `SystemWorker` nuevo (cada hora): calcula salud por cuenta y snapshot del sistema.
  - Salud: semáforo activo/en-riesgo/inactivo por tenant según dias sin login y uso clínico.
  - Sistema: estado de BD, backup, worker de métricas, disco.
- `health_handlers.go`: endpoints `GET /admin/health/accounts` y `GET /admin/system`.
- `auth_handlers.go`: registra intentos de login fallidos en `login_attempts`.
- Frontend: pestañas nuevas "Salud" y "Sistema" en panel admin.
  - Salud: tabla con filtros por semáforo, antigüedad, sesiones/mes, notas/mes, módulos.
  - Sistema: 4 tarjetas de estado + tabla de últimos accesos fallidos.

### Archivos involucrados
- `database/migrations/000028_health_system_snapshot.up.sql` (nuevo)
- `internal/workers/system_worker.go` (nuevo)
- `internal/delivery/http/api/health_handlers.go` (nuevo)
- `internal/delivery/http/api/auth_handlers.go`
- `internal/delivery/http/api/deps.go`
- `internal/delivery/http/api/router.go`
- `cmd/vuhmik-api/main.go`
- `frontend/src/presentation/views/AdminView.vue`

---

## Issue #241 — Métricas: grupos por plan, actividad poblada, fix activity_handlers

**Rama:** `issue/241-metricas-grupos-por-plan`
**Merge a main:** `995eea1`
**Capa:** Workers / Delivery / Asteroide (frontend).

### Problema
1. El "Detalle por cuenta" en Métricas mostraba todos los tenants mezclados sin
   distinguir entre plan mensual y por módulo.
2. El panel de Actividad siempre mostraba vacío — `activity_snapshot` nunca se
   poblaba porque ningún worker lo calculaba.
3. El drill-down mensual de Actividad devolvía `periods: []` por el mismo problema
   de tipo `date` vs `string` de pgx v5.

### Solución
- `MetricsWorker`: agrega `billing_mode` y `monthly_fee` al struct `accountDetail`
  y al query de `accounts_detail` para que el frontend pueda agrupar por tipo de plan.
- `SystemWorker`: nuevo método `calculateActivitySnapshot()` — calcula conteos
  mensuales por tenant desde las proyecciones y los guarda en `activity_snapshot`.
- `activity_handlers.go`: cast `period::text` y `MAX(s.period)::text` para
  compatibilidad con pgx v5 (mismo fix que se aplicó en `metrics_handlers.go`).
- Frontend: "Detalle por cuenta" en Métricas muestra grupos visuales — título
  "Plan mensual" con subtítulos por precio (ej. $2,500/mes), y "Por módulo" debajo.

### Archivos involucrados
- `internal/workers/system_worker.go`
- `internal/workers/metrics_worker.go`
- `internal/delivery/http/api/activity_handlers.go`
- `frontend/src/presentation/views/AdminView.vue`

---

