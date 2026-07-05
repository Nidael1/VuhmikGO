# ADR-0022 — CQRS: proyecciones de lectura por Shader

## Estado
Aceptado

## Fecha
2026-06-27

## Contexto

La tabla `evidence` es el registro de auditoría del Core — append-only,
inmutable, agnóstica. Hoy los Services (AllergyService, etc.) consultan
evidencia haciendo FindAll(tenantID) y filtrando en memoria por
subject_ref y content->>'type'. Esto funciona correctamente para un
médico en piloto pero se degrada linealmente con el volumen:

  - Un tenant con 10,000 registros de evidencia carga todos en memoria
    para devolver 20 alergias de un paciente.
  - El servidor VPS hace el trabajo que debería hacer la BD.
  - No escala a múltiples médicos simultáneos sin aumentar RAM/CPU.

El objetivo es optimizar el servidor VPS al máximo: que la BD haga el
trabajo de filtrado con índices, y que el cliente web haga el trabajo
de presentación. El servidor solo valida, escribe y sirve.

## Decisión

### Patrón CQRS aplicado a VUHMÍK

Separar escritura de lectura en la capa de Shaders:

  WRITE (comando) → Core: evidence (inmutable, fuente de verdad legal)
  READ  (query)   → Shader: tablas de proyección optimizadas para consulta

El Core no cambia. Las proyecciones son responsabilidad del Shader que
conoce el tipo de contenido. Si hay discrepancia entre evidence y la
proyección, evidence siempre gana.

### Tablas de proyección (capa Shader, no Core)

Cada módulo clínico tiene su tabla de proyección:

  note_projections
    evidence_id    TEXT PK FK → evidence.id
    tenant_id      TEXT NOT NULL
    patient_id     TEXT NOT NULL
    text           TEXT NOT NULL
    created_at     TIMESTAMPTZ NOT NULL
    issued_at      TIMESTAMPTZ
    state          TEXT NOT NULL

  allergy_projections
    evidence_id    TEXT PK FK → evidence.id
    tenant_id      TEXT NOT NULL
    patient_id     TEXT NOT NULL
    agente         TEXT NOT NULL
    tipo_reaccion  TEXT NOT NULL
    criticidad     TEXT
    certeza        TEXT
    fecha_inicio   TEXT
    notas          TEXT
    state          TEXT NOT NULL
    created_at     TIMESTAMPTZ NOT NULL

  prescription_projections
    evidence_id          TEXT PK FK → evidence.id
    tenant_id            TEXT NOT NULL
    patient_id           TEXT NOT NULL
    medicamento_generico TEXT NOT NULL
    dosis                TEXT NOT NULL
    diagnostico          TEXT
    indicaciones         TEXT
    seguimiento          TEXT
    state                TEXT NOT NULL
    created_at           TIMESTAMPTZ NOT NULL
    issued_at            TIMESTAMPTZ

### Índices de proyecciones

Cada tabla de proyección tiene índices por tenant_id + patient_id
y por state para filtrar activos/voided eficientemente:

  idx_note_proj_patient        (tenant_id, patient_id)
  idx_allergy_proj_patient     (tenant_id, patient_id)
  idx_allergy_proj_state       (tenant_id, state)
  idx_prescription_proj_patient (tenant_id, patient_id)
  idx_prescription_proj_state   (tenant_id, state)

### Índice en evidence para reconstrucción

Para poder reconstruir proyecciones desde evidence si es necesario:

  idx_evidence_content_type    ((content->>'type'))

### Flujo de escritura (Shader)

Toda escritura es atómica — dentro de la misma transacción PostgreSQL:

  1. Core.Create(evidence) — registro inmutable
  2. ProjectionRepo.Upsert(projection) — fila de lectura rápida
  3. Si cualquiera falla → rollback total

### Flujo de lectura (Shader)

Los Services leen SOLO de las tablas de proyección, nunca de evidence
para queries de lista. Evidence solo se lee para:
  - Export legal (necesita el blob completo con hash)
  - Void+replace (necesita el registro original)
  - Reconstrucción de proyección si hay inconsistencia

### Trabajo en el cliente web

Las proyecciones devuelven exactamente los campos que necesita el
frontend — sin parseo de JSON en el servidor. El cliente recibe datos
planos y hace el trabajo de presentación (ordenar, filtrar en UI,
paginar en pantalla).

### Métricas y paneles (ADR-0019, ADR-0023)

Las tablas de proyección son también la fuente de los conteos para
los paneles de métricas y actividad:

  COUNT(*) FROM allergy_projections WHERE tenant_id = X AND state = 'issued'
  COUNT(*) FROM prescription_projections WHERE tenant_id = X

Sin tocar evidence. Sin PHI en las métricas.

## Dependencias

  - ADR-0016: el Core sigue agnóstico — las proyecciones viven en
              la capa Shader, no en el Core.
  - ADR-0006: void+replace actualiza tanto evidence como la proyección
              en la misma transacción.
  - ADR-0008: el hash se calcula sobre evidence (blob completo),
              no sobre la proyección.
  - ADR-0011: prescription_projections es la tabla de lectura rápida
              para el módulo de receta electrónica.
  - ADR-0019: los paneles de métricas leen de proyecciones, no de evidence.
  - ADR-0023: el panel de actividad cuenta registros en proyecciones.

## Estado de implementación

  Implementado. Migración 000011_projections.up.sql aplicada.
  NoteProjectionRepository, PrescriptionProjectionRepository y
  ConsultationProjectionRepository implementados en ports/ y postgres/.
  Ejecutado en el desarrollo post-MVP previo a esta sesión.
    - Tabla prescription_projections + índices

  Refactor AllergyService:
    - Create: escritura atómica evidence + allergy_projections
    - ListByPatient: lee de allergy_projections
    - Void: actualiza estado en allergy_projections

  Refactor ECEService (notas):
    - Draft+emit: escribe en note_projections
    - Edit (void+replace): actualiza note_projections
    - List: lee de note_projections

  PrescriptionService (nuevo):
    - Create: escritura atómica evidence + prescription_projections
    - List: lee de prescription_projections
    - Emit: actualiza state en prescription_projections
    - Void: actualiza state en prescription_projections

  Puerto ProjectionRepository por módulo.

## Consecuencias

  Las queries de lista son O(log n) en vez de O(n) — índices en columnas
  nativas de PostgreSQL, no dentro de JSON.
  El servidor VPS hace mínimo trabajo en lecturas — la BD filtra,
  el cliente presenta.
  Las proyecciones son reconstruibles desde evidence en cualquier momento
  — la fuente de verdad legal nunca se pierde.
  Void+replace requiere transacción explícita — agrega complejidad en
  el Service pero es manejable y correcto.
  Los paneles de métricas y actividad tienen fuente de datos eficiente
  sin tocar PHI ni evidence directamente.
