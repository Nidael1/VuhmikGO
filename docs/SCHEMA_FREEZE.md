# SCHEMA FREEZE — Issue #63

## Fecha
2026-06-12

## Version congelada
v0.1.0-rc1

## Declaracion

El esquema de base de datos del MVP queda formalmente congelado.
No se agregan migraciones nuevas ni se alteran columnas, tipos
o constraints sin un nuevo ciclo de planificacion.

## Migraciones vigentes

  000001_create_evidence.up.sql
  000002_create_evidence_indexes.up.sql

  Aplicadas via golang-migrate. Versionadas en schema_migrations.
  Sin archivos .down.sql (migraciones forward-only por diseno).

## Schema congelado — tabla evidence

  Columnas:
    id             TEXT          NOT NULL (PK)
    tenant_id      TEXT          NOT NULL
    state          TEXT          NOT NULL
    created_at     TIMESTAMPTZ   NOT NULL
    issued_at      TIMESTAMPTZ   NULL
    voided_at      TIMESTAMPTZ   NULL
    replaced_by_id TEXT          NULL

  Constraints:
    evidence_pkey              PRIMARY KEY (id)
    evidence_state_valid       CHECK state IN (draft,issued,locked,voided)
    evidence_tenant_id_not_empty CHECK tenant_id != ''
    evidence_replaced_by_fk    FK replaced_by_id -> evidence(id)
                               ON UPDATE RESTRICT ON DELETE RESTRICT

  Indices:
    idx_evidence_tenant_id     btree (tenant_id)
    idx_evidence_state         btree (state)
    idx_evidence_issued_at     btree (issued_at)

## Reglas del freeze

1. No se agregan tablas sin ADR y migracion forward-only.
2. No se alteran columnas, tipos ni constraints existentes.
3. No se eliminan indices existentes.
4. No se crean archivos .down.sql.
5. Toda nueva migracion debe ser numerada secuencialmente
   (000003, 000004, ...) y aprobada via ADR.
