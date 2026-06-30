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

## Issue B — Navegación: detalle de instancia vs. perfil de paciente

**Rama:** `issue/crm-prescription-consultation-detail-view` *(pendiente de crear)*
**Commit:** *(pendiente)*
**Capa:** Asteroide `crm_ui` (UX/Producto). No toca Shaders ni Core.

### Problema
Al abrir una receta o consulta desde su sección propia en el menú lateral (`Recetas`, `Consultas`), el comportamiento actual redirige al perfil completo del paciente (`/patients/{id}`) en lugar de mostrar el detalle de esa instancia específica (la receta o la consulta individual).

### Solución propuesta
*(pendiente de ejecución)*

### Archivos involucrados (identificados, pendientes de modificar)
- `frontend/src/presentation/views/PrescriptionListView.vue`
- `frontend/src/presentation/views/ConsultationListView.vue`
- Posible vista nueva de detalle de receta y/o consulta (a definir)

### Estado
🔲 No iniciado.

---

## Notas de ejecución

Durante la ejecución de Issue A se identificó una desincronización entre dos copias locales del repositorio:
- `/Volumes/D/vuhmikGo` — repositorio real, conectado a `origin` (GitHub: `Nidael1/VuhmikGO`).
- `/Volumes/D/Copia de VuhmikGO` — copia de respaldo sin `.git` propio, usada temporalmente como espacio de trabajo.

Se identificó trabajo de backend pendiente sin commitear en `main` (prescripción vinculada a `consultation_id`, migración `000017_prescription_consultation_id.up.sql`, handlers y router relacionados). Este trabajo se resguardó mediante:

```
git stash push -m "wip-issue-156-prescription-consultation-link"
```

Pendiente de asignar a su propio issue/rama (siguiendo la convención de numeración interna del proyecto, correspondería al **issue #156**, siguiente al último mergeado en `main`, issue #155).
