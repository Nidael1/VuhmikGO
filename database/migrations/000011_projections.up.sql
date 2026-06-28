-- ADR-0022: CQRS — tablas de proyección de lectura por Shader
-- Las proyecciones son responsabilidad del Shader, no del Core.
-- El Core (evidence) sigue siendo la fuente de verdad legal.
-- Si hay discrepancia, evidence gana siempre.

-- Índice en evidence para reconstrucción y filtrado por tipo
CREATE INDEX IF NOT EXISTS idx_evidence_content_type
    ON evidence ((content->>'type'));

-- Proyección de notas clínicas
CREATE TABLE IF NOT EXISTS note_projections (
    evidence_id  TEXT        NOT NULL REFERENCES evidence(id) ON DELETE CASCADE,
    tenant_id    TEXT        NOT NULL,
    patient_id   TEXT        NOT NULL,
    text         TEXT        NOT NULL DEFAULT '',
    state        TEXT        NOT NULL DEFAULT 'draft',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    issued_at    TIMESTAMPTZ,
    PRIMARY KEY (evidence_id)
);

CREATE INDEX IF NOT EXISTS idx_note_proj_patient
    ON note_projections (tenant_id, patient_id);

CREATE INDEX IF NOT EXISTS idx_note_proj_state
    ON note_projections (tenant_id, state);

-- Proyección de alergias
CREATE TABLE IF NOT EXISTS allergy_projections (
    evidence_id   TEXT        NOT NULL REFERENCES evidence(id) ON DELETE CASCADE,
    tenant_id     TEXT        NOT NULL,
    patient_id    TEXT        NOT NULL,
    agente        TEXT        NOT NULL DEFAULT '',
    tipo_reaccion TEXT        NOT NULL DEFAULT '',
    criticidad    TEXT,
    certeza       TEXT,
    fecha_inicio  TEXT,
    notas         TEXT,
    state         TEXT        NOT NULL DEFAULT 'draft',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    issued_at     TIMESTAMPTZ,
    PRIMARY KEY (evidence_id)
);

CREATE INDEX IF NOT EXISTS idx_allergy_proj_patient
    ON allergy_projections (tenant_id, patient_id);

CREATE INDEX IF NOT EXISTS idx_allergy_proj_state
    ON allergy_projections (tenant_id, state);

-- Proyección de recetas electrónicas
CREATE TABLE IF NOT EXISTS prescription_projections (
    evidence_id          TEXT        NOT NULL REFERENCES evidence(id) ON DELETE CASCADE,
    tenant_id            TEXT        NOT NULL,
    patient_id           TEXT        NOT NULL,
    medicamento_generico TEXT        NOT NULL DEFAULT '',
    dosis                TEXT        NOT NULL DEFAULT '',
    diagnostico          TEXT,
    indicaciones         TEXT,
    seguimiento          TEXT,
    state                TEXT        NOT NULL DEFAULT 'draft',
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    issued_at            TIMESTAMPTZ,
    PRIMARY KEY (evidence_id)
);

CREATE INDEX IF NOT EXISTS idx_prescription_proj_patient
    ON prescription_projections (tenant_id, patient_id);

CREATE INDEX IF NOT EXISTS idx_prescription_proj_state
    ON prescription_projections (tenant_id, state);
