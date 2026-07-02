# ADR-0026 — Layout de dos columnas en el expediente del paciente

## Estado
Aceptado

## Fecha
2026-07-02

---

## Contexto

El expediente del paciente (Asteroide `crm_ui`) se presenta actualmente en una sola columna, con secciones colapsables cerradas por defecto (issue #170): alergias, recetas, consultas y notas clínicas.

Las notas clínicas siguen el modelo void+replace definido en ADR-0006 y se incorporaron al expediente en el issue #169.

El modelo de datos ya distingue notas generales de notas por consulta: la migración 000016 agregó `consultation_id` a `note_projections`, por lo que una nota puede o no estar asociada a una consulta específica.

---

## Problema

Las notas clínicas son el contenido que el médico necesita consultar con mayor frecuencia durante la atención. En el layout de columna única, quedan colapsadas junto con el resto de las secciones, obligando a expandir y colapsar repetidamente mientras se revisa el resto del expediente (alergias, recetas, consultas). Esta es la corrección UX pendiente que bloquea el issue #171.

---

## Decisión

Se adopta un layout de dos columnas para el expediente del paciente:

- **Columna principal (65%, izquierda):** datos del paciente, alergias, recetas (`rx-card`) y consultas (`con-card`) — contenido organizado cronológicamente por fecha/consulta. Conserva el comportamiento colapsable ya implementado en el issue #170.
- **Panel lateral (35%, derecha):** notas generales del paciente — notas con `consultation_id = NULL`, no atadas a una consulta ni a un momento específico. Fuera del listado de secciones colapsables, visible sin necesidad de expandir nada.
  - Las notas que sí tienen `consultation_id` (creadas dentro de una consulta) no se duplican aquí; se consultan desde el detalle de la consulta correspondiente (`/consultations/:id`, issues #172 y #174).
- **Desktop:** ambas columnas conviven lado a lado. El panel lateral permanece visible (`position: sticky`) durante el scroll de la columna principal, porque su contenido es contexto general y persistente del paciente — no un evento puntual — y debe seguir visible mientras se recorre el historial cronológico de la izquierda.
  - **Doble clic en el panel derecho:** contrae el panel de notas; la columna izquierda pasa a ocupar el 100% del ancho. Un segundo doble clic lo expande de vuelta a 65/35. El panel izquierdo nunca se contrae; siempre está visible.
  - El estado contraído/expandido no se persiste entre sesiones ni entre pacientes; al abrir un expediente se muestra siempre en 65/35 por defecto.
- **Móvil:** el layout se apila en una sola columna; sin sticky (no aplica sin dos columnas lado a lado).

El breakpoint exacto desktop/móvil no está definido en la documentación cargada; se deja como detalle de implementación del issue #171. Si las notas generales llegan a superar el alto del viewport, el panel necesita scroll interno propio para no cortar contenido.

---

## Alternativas consideradas

### Mantener columna única con notas como primera sección
Rechazada: no resuelve el problema, solo cambia el orden; las notas seguirían ocultas por defecto.

### Modal o drawer para notas clínicas
Rechazada: exige una acción extra (abrir/cerrar) y no permite ver notas y expediente al mismo tiempo, que es justamente el objetivo.

### Pestañas (tabs) en lugar de panel lateral
Rechazada: oculta el resto del expediente mientras se ven las notas; el panel lateral permite consulta simultánea.

### Split 50/50 en vez de 65/35
Rechazada: el expediente principal (datos, alergias, recetas, consultas) requiere más espacio horizontal que el listado de notas.

### Mostrar todas las notas (generales y por consulta) en el panel lateral
Rechazada: duplicaría las notas de consulta en dos lugares (panel lateral y detalle de consulta) y diluiría la distinción entre contexto general del paciente e historial cronológico por evento que motiva el layout de dos columnas.

---

## Consecuencias

- Cambio exclusivo de presentación en el Asteroide `crm_ui`. No toca Core, Shaders, modelos de datos ni endpoints.
- Las notas clínicas salen del listado de secciones colapsables del issue #170; el resto de secciones no cambia.
- No introduce funcionalidad nueva de edición sobre las notas: creación/edición/void+replace sigue gobernada por ADR-0006 sin cambios.
- El filtrado a "solo notas generales" (`consultation_id = NULL`) se resuelve en frontend si el endpoint de notas ya expone ese campo por nota (esperable, existe en el modelo desde la migración 000016). Si no lo expone todavía, se necesita un ajuste menor de backend antes de ejecutar #171.
- Requiere definir el breakpoint responsive durante la implementación del issue #171.

---

## Documentos impactados

- `VUHMIK_ESTADO_ACTUAL.md` — mover ADR-0026 de "pendientes de escribir" a "ADRs existentes relevantes" y desbloquear el issue #171.
- `ASTEROID_crm_ui.md` — este ADR precisa el layout de la pantalla "Vista de expediente del paciente" ya listada en ese documento.

---

## Regla final

Este ADR autoriza únicamente el layout de dos columnas (65/35) del expediente del paciente, la extracción de las notas generales del paciente (`consultation_id = NULL`) del listado colapsable hacia un panel lateral sticky en desktop / apilado en móvil, el toggle de contracción/expansión del panel derecho mediante doble clic, y el filtrado correspondiente para excluir notas por consulta de ese panel.

No autoriza cambios en el modelo de datos, en el lifecycle de evidencia clínica, en Shaders, en Core, ni funcionalidad nueva de edición o interacción sobre notas clínicas más allá de la ya existente.
