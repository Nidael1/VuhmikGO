-- Core: registros de evidencia
-- Migración: solo hacia adelante — sin rollback en producción
-- Campos: los 7 campos Core definidos en Issue #1
-- ADR-0001: PostgreSQL, migraciones forward-only

CREATE TABLE IF NOT EXISTS evidence (
    id             TEXT        NOT NULL,
    tenant_id      TEXT        NOT NULL,
    state          TEXT        NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL,
    issued_at      TIMESTAMPTZ,
    voided_at      TIMESTAMPTZ,
    replaced_by_id TEXT,

    CONSTRAINT evidence_pkey
        PRIMARY KEY (id),

    CONSTRAINT evidence_tenant_id_not_empty
        CHECK (tenant_id <> ''),

    CONSTRAINT evidence_state_valid
        CHECK (state IN ('draft', 'issued', 'locked', 'voided')),

    CONSTRAINT evidence_replaced_by_fk
        FOREIGN KEY (replaced_by_id)
        REFERENCES evidence (id)
        ON DELETE RESTRICT
        ON UPDATE RESTRICT
);
