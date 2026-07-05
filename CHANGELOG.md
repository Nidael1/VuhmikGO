# CHANGELOG — VUHMÍK

Registro de issues ejecutados sobre el Asteroide `crm_ui`. Cada entrada corresponde a un issue independiente: una rama, un commit, un PR.

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
