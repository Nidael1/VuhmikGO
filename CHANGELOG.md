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

