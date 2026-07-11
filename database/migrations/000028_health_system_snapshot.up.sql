-- Migración 000028: health_snapshot — salud de cuentas y estado del sistema
CREATE TABLE IF NOT EXISTS account_health_snapshot (
    tenant_id               TEXT        NOT NULL,
    email                   TEXT        NOT NULL,
    calculated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    account_age_days        INTEGER     NOT NULL DEFAULT 0,
    last_login_at           TIMESTAMPTZ,
    days_since_login        INTEGER     NOT NULL DEFAULT 0,
    sessions_this_month     INTEGER     NOT NULL DEFAULT 0,
    sessions_last_month     INTEGER     NOT NULL DEFAULT 0,
    notes_this_month        INTEGER     NOT NULL DEFAULT 0,
    prescriptions_this_month INTEGER    NOT NULL DEFAULT 0,
    allergies_this_month    INTEGER     NOT NULL DEFAULT 0,
    total_notes             INTEGER     NOT NULL DEFAULT 0,
    total_prescriptions     INTEGER     NOT NULL DEFAULT 0,
    total_patients          INTEGER     NOT NULL DEFAULT 0,
    modules_active          INTEGER     NOT NULL DEFAULT 0,
    modules_used            INTEGER     NOT NULL DEFAULT 0,
    health_status           TEXT        NOT NULL DEFAULT 'active',
    PRIMARY KEY (tenant_id)
);

CREATE TABLE IF NOT EXISTS system_snapshot (
    id                  TEXT        PRIMARY KEY DEFAULT gen_random_uuid()::text,
    calculated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    db_ok               BOOLEAN     NOT NULL DEFAULT true,
    last_backup_at      TIMESTAMPTZ,
    last_backup_size_kb INTEGER     NOT NULL DEFAULT 0,
    backup_ok           BOOLEAN     NOT NULL DEFAULT true,
    metrics_last_run_at TIMESTAMPTZ,
    metrics_ok          BOOLEAN     NOT NULL DEFAULT true,
    disk_used_pct       INTEGER     NOT NULL DEFAULT 0,
    disk_ok             BOOLEAN     NOT NULL DEFAULT true,
    overall_ok          BOOLEAN     NOT NULL DEFAULT true,
    issues              TEXT        NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_system_snapshot_calculated_at
    ON system_snapshot (calculated_at DESC);

CREATE TABLE IF NOT EXISTS login_attempts (
    id          TEXT        PRIMARY KEY DEFAULT gen_random_uuid()::text,
    email       TEXT        NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reason      TEXT        NOT NULL DEFAULT 'invalid_credentials'
);

CREATE INDEX IF NOT EXISTS idx_login_attempts_recent
    ON login_attempts (occurred_at DESC);
