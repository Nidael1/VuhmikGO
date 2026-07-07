-- ADR-0023: Panel de actividad y uso.
-- Snapshot precalculado por tenant y periodo (mes).
-- Separado de metrics_snapshot (ADR-0019) que es global.
-- Migracion forward-only. Sin .down.sql.

CREATE TABLE IF NOT EXISTS activity_snapshot (
    tenant_id           TEXT    NOT NULL,
    period              DATE    NOT NULL,
    notes_count         INTEGER NOT NULL DEFAULT 0,
    allergies_count     INTEGER NOT NULL DEFAULT 0,
    prescriptions_count INTEGER NOT NULL DEFAULT 0,
    exports_count       INTEGER NOT NULL DEFAULT 0,
    patients_count      INTEGER NOT NULL DEFAULT 0,
    sessions_count      INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (tenant_id, period)
);

CREATE INDEX IF NOT EXISTS idx_activity_snapshot_tenant
    ON activity_snapshot (tenant_id, period DESC);

CREATE INDEX IF NOT EXISTS idx_activity_snapshot_period
    ON activity_snapshot (period DESC);
