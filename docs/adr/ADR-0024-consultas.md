# ADR-0024 — Módulo de consulta médica

## Estado
Propuesto

## Fecha
2026-06-28

## Contexto

El flujo real de una consulta médica tiene tres partes vinculadas:

  1. Signos vitales — exploración física (T/A, FC, FR, Temp, Peso, Talla, SAO2).
  2. Nota clínica — observaciones del médico (anamnesis, exploración,
     impresión diagnóstica). No necesariamente implica medicación.
  3. Receta — prescripción de medicamentos. Opcional. Puede no haber.

En la implementación anterior (Sprint 9.3-9.4) estos tres elementos
viven en tablas separadas sin vínculo explícito entre sí:
  - Los signos vitales están en note_projections como campos de la nota.
  - La receta no tiene referencia a la nota que la originó.
  - No existe el concepto de "consulta" como entidad.

Esto genera problemas clínicos y de UX:
  - El médico no puede ver la consulta completa en un solo lugar.
  - El PDF de receta no puede jalar signos vitales de la nota.
  - El expediente muestra notas y recetas desvinculadas.
  - No se puede saber qué receta se emitió en qué consulta.

La NOM-004-SSA3-2012 define la consulta como la unidad básica del
expediente clínico. La nota, los signos vitales y la receta son
componentes de la consulta, no entidades independientes.

## Decisión

### La consulta como entidad de primer nivel

Una consulta (Consultation) es la unidad básica del expediente.
Agrupa signos vitales, nota clínica y receta opcional bajo un
identificador único (consultation_id).

### Tabla de proyección consultation_projections

La consulta sigue el patrón CQRS (ADR-0022):
  - El Core sigue siendo agnóstico — una consulta es un registro
    de evidencia con type: "consultation" en el blob.
  - El Shader proyecta a consultation_projections para lectura rápida.

  consultation_projections:
    evidence_id     TEXT PK FK → evidence.id
    tenant_id       TEXT NOT NULL
    patient_id      TEXT NOT NULL
    -- Signos vitales (opcionales)
    ta              TEXT NOT NULL DEFAULT ''
    fc              TEXT NOT NULL DEFAULT ''
    fr              TEXT NOT NULL DEFAULT ''
    temp            TEXT NOT NULL DEFAULT ''
    peso            TEXT NOT NULL DEFAULT ''
    talla           TEXT NOT NULL DEFAULT ''
    sao2            TEXT NOT NULL DEFAULT ''
    -- Estado
    state           TEXT NOT NULL DEFAULT 'draft'
    created_at      TIMESTAMPTZ NOT NULL
    issued_at       TIMESTAMPTZ

### Vinculación de nota y receta a la consulta

  note_projections:
    ADD COLUMN consultation_id TEXT  (FK opcional → consultation_projections)

  prescription_projections:
    consultation_id ya existe (migración 000015)
    Se actualiza para referenciar consultation_projections.

### Flujo de una consulta

  1. Médico abre "Nueva consulta" para un paciente.
  2. Sistema crea un borrador de consulta (draft).
  3. Médico llena signos vitales (opcionales).
  4. Médico escribe la nota clínica — se vincula al consultation_id.
  5. Médico opcionalmente emite una receta — se vincula al consultation_id.
  6. Médico emite la consulta — queda issued, inmutable.
  7. El PDF de receta jala signos vitales de la consulta vinculada.

### Navegación

  Sidebar: Pacientes · Consultas · Recetas · Mi perfil

  Desde Consultas:
    - Lista global de consultas del tenant (más reciente primero)
    - Nueva consulta (selecciona paciente)
    - Detalle de consulta: signos vitales + nota + receta

  Desde Pacientes (detalle del paciente):
    - Historial de consultas del paciente (reemplaza "Expediente clínico")
    - Cada consulta muestra: fecha, signos vitales, nota, receta si existe

  Desde Recetas:
    - La receta muestra el botón "Ver consulta" si tiene consultation_id

### Retrocompatibilidad

  Las notas y recetas existentes sin consultation_id siguen
  apareciendo en el expediente del paciente bajo "Sin consulta".
  No se migran datos históricos — solo las nuevas consultas
  usan el flujo vinculado.

### PDF de receta

  El PDF de receta incluye los signos vitales de la consulta vinculada
  si existe. Si no hay consulta vinculada, los campos de signos vitales
  aparecen en blanco para llenar a mano.

## Dependencias

  - ADR-0022: CQRS — consultation_projections sigue el mismo patrón
              que note_projections y prescription_projections.
  - ADR-0006: void+replace — correcciones de consulta siguen el mismo
              mecanismo que notas y recetas.
  - ADR-0011: la receta referencia opcionalmente una consulta.
  - ADR-0024: define el Shader de consultas (ConsultationShader).

## Estado de implementación

  No implementado.
  Requiere issues de implementación:

  Migración 000016:
    - Tabla consultation_projections + índices
    - ADD COLUMN consultation_id en note_projections
    - El consultation_id en prescription_projections ya existe

  ConsultationShader:
    - Valida campos mínimos de una consulta
    - Blob type: "consultation"

  ConsultationService:
    - CreateDraft, Emit, Void
    - LinkNote(consultationID, noteID)
    - LinkPrescription(consultationID, prescriptionID)

  API handlers:
    - POST /api/v1/patients/:id/consultations (crear)
    - GET  /api/v1/patients/:id/consultations (listar por paciente)
    - GET  /api/v1/consultations (lista global)
    - POST /api/v1/consultations/:id/emit
    - POST /api/v1/consultations/:id/void

  Frontend:
    - Vista ConsultationListView (sidebar)
    - Vista ConsultationNewView (formulario unificado)
    - Vista ConsultationDetailView
    - Sección consultas en PatientDetailView
    - Link "Ver consulta" en recetas

## Consecuencias

  El médico tiene un flujo unificado para la consulta completa.
  El PDF de receta puede incluir signos vitales automáticamente.
  El expediente es coherente — cada consulta es una unidad.
  Las notas y recetas sin consultation_id siguen funcionando.
  La arquitectura CQRS y el Core agnóstico no cambian.
