-- ADR-0024: Módulo de consulta médica
-- La consulta es la unidad básica del expediente clínico.
-- Agrupa signos vitales, nota clínica y receta opcional.

CREATE TABLE IF NOT EXISTS consultation_projections (
    evidence_id  TEXT        NOT NULL REFERENCES evidence(id) ON DELETE CASCADE,
    tenant_id    TEXT        NOT NULL,
    patient_id   TEXT        NOT NULL,
    -- Signos vitales (opcionales)
    ta           TEXT        NOT NULL DEFAULT '',
    fc           TEXT        NOT NULL DEFAULT '',
    fr           TEXT        NOT NULL DEFAULT '',
    temp         TEXT        NOT NULL DEFAULT '',
    peso         TEXT        NOT NULL DEFAULT '',
    talla        TEXT        NOT NULL DEFAULT '',
    sao2         TEXT        NOT NULL DEFAULT '',
    -- Estado
    state        TEXT        NOT NULL DEFAULT 'draft',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    issued_at    TIMESTAMPTZ,
    PRIMARY KEY (evidence_id)
);

CREATE INDEX IF NOT EXISTS idx_consultation_proj_patient
    ON consultation_projections (tenant_id, patient_id);

CREATE INDEX IF NOT EXISTS idx_consultation_proj_state
    ON consultation_projections (tenant_id, state);

-- Vincular nota clínica a consulta
ALTER TABLE note_projections
    ADD COLUMN IF NOT EXISTS consultation_id TEXT;

CREATE INDEX IF NOT EXISTS idx_note_proj_consultation
    ON note_projections (consultation_id) WHERE consultation_id IS NOT NULL;
