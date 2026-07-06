-- ADR-0015: Modulo de resultados de laboratorio.
-- El Core almacena resultados como blobs opacos en evidence (type: lab_result).
-- Esta tabla es la proyeccion CQRS de lectura rapida (ADR-0022).
-- Migracion forward-only. Sin .down.sql.

CREATE TABLE IF NOT EXISTS lab_result_projections (
    evidence_id    TEXT        NOT NULL REFERENCES evidence(id) ON DELETE CASCADE,
    tenant_id      TEXT        NOT NULL,
    patient_id     TEXT        NOT NULL,
    estudio        TEXT        NOT NULL DEFAULT '',
    fecha_estudio  TEXT        NOT NULL DEFAULT '',
    resultado      TEXT,
    laboratorio    TEXT,
    unidades       TEXT,
    valor_referencia TEXT,
    notas          TEXT,
    state          TEXT        NOT NULL DEFAULT 'draft',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    issued_at      TIMESTAMPTZ,
    PRIMARY KEY (evidence_id)
);

CREATE INDEX IF NOT EXISTS idx_lab_result_proj_patient
    ON lab_result_projections (tenant_id, patient_id);

CREATE INDEX IF NOT EXISTS idx_lab_result_proj_state
    ON lab_result_projections (tenant_id, state);
