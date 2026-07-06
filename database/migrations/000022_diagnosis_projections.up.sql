-- ADR-0013: Modulo de diagnosticos estructurados y lista de problemas (CIE-10).
-- El Core almacena diagnosticos como blobs opacos en evidence (type: diagnosis).
-- Esta tabla es la proyeccion CQRS de lectura rapida (ADR-0022).
-- Migracion forward-only. Sin .down.sql.

CREATE TABLE IF NOT EXISTS diagnosis_projections (
    evidence_id     TEXT        NOT NULL REFERENCES evidence(id) ON DELETE CASCADE,
    tenant_id       TEXT        NOT NULL,
    patient_id      TEXT        NOT NULL,
    descripcion     TEXT        NOT NULL DEFAULT '',
    codigo_cie10    TEXT,
    tipo            TEXT,
    estado_problema TEXT,
    fecha_inicio    TEXT,
    notas           TEXT,
    state           TEXT        NOT NULL DEFAULT 'draft',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    issued_at       TIMESTAMPTZ,
    PRIMARY KEY (evidence_id)
);

CREATE INDEX IF NOT EXISTS idx_diagnosis_proj_patient
    ON diagnosis_projections (tenant_id, patient_id);

CREATE INDEX IF NOT EXISTS idx_diagnosis_proj_state
    ON diagnosis_projections (tenant_id, state);
