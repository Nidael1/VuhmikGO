# ADR-0023 — Panel de actividad y uso

## Estado
Aceptado

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

  1. Tabla activity_snapshot (este ADR):
       conteos clínicos por tenant y por periodo (mes)
       sin tocar evidence, sin PHI

  2. Tabla activity_log (migración 000013):
       registra eventos de acceso al sistema
       sin datos clínicos, solo metadatos de uso

### Tabla activity_log

Registra eventos de uso del sistema sin PHI:

  activity_log:
    id          TEXT PK
    tenant_id   TEXT NOT NULL
    event_type  TEXT NOT NULL
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW()

  Tipos de evento implementados en v1:
    session_start    — médico inicia sesión (HandleLogin)
    session_end      — médico cierra sesión (HandleLogout)

  Índices:
    idx_activity_tenant_date (tenant_id, occurred_at)
    idx_activity_event_type  (event_type, occurred_at)

### Tabla activity_snapshot

Snapshot precalculado por tenant y periodo (mes).
Separado de metrics_snapshot (ADR-0019) que es global.
PK compuesta (tenant_id, period) para evitar colisión con ADR-0019.

  activity_snapshot:
    tenant_id           TEXT NOT NULL
    period              DATE NOT NULL
    notes_count         INTEGER NOT NULL DEFAULT 0
    allergies_count     INTEGER NOT NULL DEFAULT 0
    prescriptions_count INTEGER NOT NULL DEFAULT 0
    exports_count       INTEGER NOT NULL DEFAULT 0
    patients_count      INTEGER NOT NULL DEFAULT 0
    sessions_count      INTEGER NOT NULL DEFAULT 0
    PRIMARY KEY (tenant_id, period)

### API del panel

  GET /api/v1/admin/activity          lista de médicos con conteos
  GET /api/v1/admin/activity/:tenant  detalle por mes (últimos 12)

Todos requieren is_admin = true en el JWT.
Todos son GET — solo lectura, sin mutaciones.

### Separación estricta de datos

  activity_log      → metadatos de uso, sin PHI
  activity_snapshot → conteos agregados por tenant/mes, sin PHI

El panel admin NUNCA accede a:
  - evidence (blob clínico)
  - patients (datos personales)
  - professional_profiles (datos del médico)

### Frontend

El panel de actividad se integra como sección dentro de AdminView.vue
junto con Operaciones (ADR-0018) y Métricas (ADR-0019). Navegación
lateral entre las tres secciones. Sin rutas adicionales en el router.

## Dependencias

  - ADR-0017: tenant_capabilities como fuente de plan y costo.
  - ADR-0018: comparte la bandera is_admin para acceso.
  - ADR-0019: metrics_snapshot global es independiente;
              activity_snapshot es por tenant/periodo.

## Estado de implementacion

  Implementado. Migraciones 000013 (activity_log) y
  000026 (activity_snapshot), activity_log.go helper,
  activity_handlers.go, AdminView.vue (sección Actividad).
  Issues #229, #230, #231, #233.
    - Migración 000013: tabla activity_log + índices.
    - Migración 000026: tabla activity_snapshot por
      tenant y periodo + índices.
    - Helper logActivity: registra eventos sin PHI.
      Fallo no bloquea el flujo principal.
    - session_start registrado en HandleLogin.
    - session_end registrado en HandleLogout.
    - Handler GET /api/v1/admin/activity: lista de
      tenants con conteos agregados desde activity_snapshot.
    - Handler GET /api/v1/admin/activity/:tenant: detalle
      por mes de un tenant (últimos 12 meses).
    - Todas las rutas protegidas por AdminMiddleware.
    - Sin PHI en ninguna ruta.
    - Frontend: sección Actividad en AdminView.vue con
      lista de cuentas y detalle mensual por tenant.

## Consecuencias

  El panel de actividad no impacta el rendimiento del sistema
  clínico — lee de snapshots precalculados.
  Los médicos no pueden ver datos de otros médicos — todo
  filtrado por tenant_id.
  activity_log es append-only — registro histórico completo
  de uso sin posibilidad de alteración.
  La separación de activity_snapshot y metrics_snapshot evita
  colisión entre ADR-0019 y ADR-0023.
  El panel admin unifica operaciones, métricas y actividad
  en una sola vista con navegación lateral.
