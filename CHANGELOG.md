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
