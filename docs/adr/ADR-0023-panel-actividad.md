# ADR-0023 — Panel de actividad y uso

## Estado
Propuesto

## Fecha
2026-06-27

## Contexto

VUHMÍK necesita tres paneles administrativos con responsabilidades
distintas y separadas:

  ADR-0018 — Panel de operaciones (toggles):
    Activa y desactiva módulos por cuenta. Control comercial.
    Quién tiene qué encendido. Escribe en tenant_capabilities.

  ADR-0019 — Panel de métricas de negocio:
    MRR, churn, ingresos por módulo, costo por cuenta.
    El dinero. Solo lectura.

  ADR-0023 — Panel de actividad y uso (este ADR):
    Sesiones, accesos al sistema, frecuencia de uso por médico.
    Conteos clínicos: recetas emitidas, notas creadas, alergias
    registradas, pacientes activos por periodo.
    Datos operativos que alimentan decisiones y el cálculo de
    ingresos. Solo lectura. Sin PHI.

Los tres paneles son completamente independientes. Ninguno modifica
datos clínicos. Ninguno expone contenido del expediente.

## Decisión

### Fuentes de datos

El panel de actividad lee de dos fuentes exclusivamente:

  1. Tablas de proyección (ADR-0022):
       note_projections, allergy_projections, prescription_projections
       → conteos clínicos por tenant y por periodo
       → sin tocar evidence, sin PHI

  2. Tabla de actividad de sesiones (nueva: activity_log):
       → registra eventos de acceso al sistema
       → sin datos clínicos, solo metadatos de uso

### Tabla activity_log

Registra eventos de uso del sistema sin PHI:

  activity_log:
    id          TEXT PK
    tenant_id   TEXT NOT NULL
    event_type  TEXT NOT NULL
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW()

  Tipos de evento (event_type):
    session_start    — médico inicia sesión
    session_end      — médico cierra sesión
    patient_created  — nuevo paciente registrado
    note_created     — nota clínica creada
    allergy_created  — alergia registrada
    prescription_issued — receta emitida
    export_generated — expediente exportado

  Índices:
    idx_activity_tenant_date (tenant_id, occurred_at)
    idx_activity_event_type  (event_type, occurred_at)

### Qué muestra el panel

  Por médico (tenant):
    - Última sesión
    - Total de pacientes registrados
    - Notas creadas (último mes / total)
    - Alergias registradas (último mes / total)
    - Recetas emitidas (último mes / total)
    - Exports generados (último mes / total)
    - Módulos activos (de tenant_capabilities)
    - Plan y costo (de tenant_capabilities)

  Agregados globales:
    - Total de médicos activos (con sesión en últimos 30 días)
    - Total de médicos inactivos
    - Módulo más usado
    - Módulo menos usado
    - Promedio de recetas por médico por mes

  Para métricas de negocio (alimenta ADR-0019):
    - MRR = SUM(costo) FROM tenant_capabilities WHERE active = true
    - Churn = médicos sin sesión en últimos 60 días

### Precálculo por worker (WAR-A)

Los conteos no se calculan en tiempo real por request — se
precalculan por el worker de métricas cada 24 horas y se
guardan en una tabla de snapshot:

  metrics_snapshot:
    tenant_id        TEXT NOT NULL
    period           DATE NOT NULL  (primer día del mes)
    notes_count      INTEGER NOT NULL DEFAULT 0
    allergies_count  INTEGER NOT NULL DEFAULT 0
    prescriptions_count INTEGER NOT NULL DEFAULT 0
    exports_count    INTEGER NOT NULL DEFAULT 0
    patients_count   INTEGER NOT NULL DEFAULT 0
    sessions_count   INTEGER NOT NULL DEFAULT 0
    PRIMARY KEY (tenant_id, period)

El panel lee de metrics_snapshot — nunca hace COUNT en tiempo real.
El worker actualiza el snapshot del mes en curso cada 24h.

### Separación estricta de datos

  activity_log     → metadatos de uso, sin PHI, sin IDs de pacientes
  metrics_snapshot → conteos agregados, sin PHI
  proyecciones     → datos clínicos estructurados, con tenant_id

El panel admin NUNCA accede a:
  - evidence (blob clínico)
  - patients (datos personales)
  - professional_profiles (datos del médico)

### API del panel

  GET /api/v1/admin/activity          → lista de médicos con conteos
  GET /api/v1/admin/activity/:tenant  → detalle de un médico
  GET /api/v1/admin/metrics/snapshot  → snapshot agregado global

Todos requieren is_admin = true en el JWT.
Todos son GET — solo lectura, sin mutaciones.

## Dependencias

  - ADR-0017: tenant_capabilities como fuente de plan y costo.
  - ADR-0018: panel de operaciones es independiente pero comparte
              la bandera is_admin para acceso.
  - ADR-0019: metrics_snapshot alimenta el panel de negocio.
  - ADR-0022: proyecciones son la fuente de conteos clínicos.
  - WAR-A: el worker de precálculo corre sobre Redis + workers Go.

## Estado de implementación

  No implementado.
  Requiere issues de implementación:

  Migración 000012 (después de 000011 — proyecciones):
    - Tabla activity_log + índices
    - Tabla metrics_snapshot + índices

  Worker de métricas (extensión del worker existente):
    - Precálculo diario de metrics_snapshot por tenant
    - Lee de proyecciones, escribe en metrics_snapshot

  Registro de eventos:
    - session_start/end en auth_handlers
    - *_created en cada Service al crear registro
    - export_generated en HandlePatientExport

  Handlers admin:
    - GET /api/v1/admin/activity
    - GET /api/v1/admin/activity/:tenant
    - GET /api/v1/admin/metrics/snapshot
    - Middleware is_admin

  Frontend admin (Asteroide):
    - Vista de lista de médicos con conteos
    - Vista de detalle por médico
    - Vista de snapshot global

## Consecuencias

  El panel de actividad no impacta el rendimiento del sistema clínico
  — lee de snapshots precalculados, no de evidence ni de proyecciones
  en tiempo real.
  Los médicos no pueden ver datos de otros médicos — todo filtrado
  por tenant_id.
  El sistema puede crecer a cientos de médicos sin que el panel
  degrade el rendimiento del VPS.
  Los datos de negocio (MRR, churn) se derivan de la misma fuente
  que el control de acceso — sin desincronización posible.
  activity_log es append-only — registro histórico completo de uso
  sin posibilidad de alteración.
