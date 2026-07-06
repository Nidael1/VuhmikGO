-- ADR-0014: Modulo de inmunizaciones y vacunacion.
-- El Core almacena vacunas como blobs opacos en evidence (type: immunization).
-- Esta tabla es la proyeccion CQRS de lectura rapida (ADR-0022).
-- Migracion forward-only. Sin .down.sql.

CREATE TABLE IF NOT EXISTS immunization_projections (
    evidence_id      TEXT        NOT NULL REFERENCES evidence(id) ON DELETE CASCADE,
    tenant_id        TEXT        NOT NULL,
    patient_id       TEXT        NOT NULL,
    vacuna           TEXT        NOT NULL DEFAULT '',
    fecha_aplicacion TEXT        NOT NULL DEFAULT '',
    lote             TEXT,
    dosis            TEXT,
    via              TEXT,
    aplicada_por     TEXT,
    notas            TEXT,
    state            TEXT        NOT NULL DEFAULT 'draft',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    issued_at        TIMESTAMPTZ,
    PRIMARY KEY (evidence_id)
);

CREATE INDEX IF NOT EXISTS idx_immunization_proj_patient
    ON immunization_projections (tenant_id, patient_id);

CREATE INDEX IF NOT EXISTS idx_immunization_proj_state
    ON immunization_projections (tenant_id, state);
