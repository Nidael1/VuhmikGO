-- ADR-0019: Panel de métricas — snapshot precalculado por worker
-- El worker calcula periódicamente y escribe aquí.
-- El panel admin solo lee, nunca calcula en vivo.

CREATE TABLE IF NOT EXISTS metrics_snapshot (
    id                   TEXT          PRIMARY KEY DEFAULT gen_random_uuid()::text,
    calculated_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    total_accounts       INTEGER       NOT NULL DEFAULT 0,
    active_accounts      INTEGER       NOT NULL DEFAULT 0,
    suspended_accounts   INTEGER       NOT NULL DEFAULT 0,
    mrr                  NUMERIC(10,2) NOT NULL DEFAULT 0,
    total_patients       INTEGER       NOT NULL DEFAULT 0,
    total_notes          INTEGER       NOT NULL DEFAULT 0,
    total_allergies      INTEGER       NOT NULL DEFAULT 0,
    total_prescriptions  INTEGER       NOT NULL DEFAULT 0,
    accounts_detail      JSONB         NOT NULL DEFAULT '[]',
    modules_distribution JSONB         NOT NULL DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_metrics_snapshot_calculated_at
    ON metrics_snapshot (calculated_at DESC);

-- ADR-0023: Registro de actividad de sesiones
CREATE TABLE IF NOT EXISTS activity_log (
    id          TEXT        PRIMARY KEY,
    tenant_id   TEXT        NOT NULL,
    event_type  TEXT        NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_activity_tenant_date
    ON activity_log (tenant_id, occurred_at DESC);

CREATE INDEX IF NOT EXISTS idx_activity_event_type
    ON activity_log (event_type, occurred_at DESC);
