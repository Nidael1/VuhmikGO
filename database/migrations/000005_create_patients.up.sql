-- Tabla de pacientes del consultorio.
-- Pertenece a los Asteroides (CRM) — no al Core.
-- El Core solo referencia al paciente por subject_id (opaco).
-- Campos minimos obligatorios segun NOM-004-SSA3-2012, numeral 5.9.
-- Migracion forward-only. Sin .down.sql.

CREATE TABLE patients (
    id              TEXT        NOT NULL,
    tenant_id       TEXT        NOT NULL,
    nombre          TEXT        NOT NULL,
    fecha_nacimiento DATE        NOT NULL,
    sexo            TEXT        NOT NULL,
    num_expediente  TEXT        NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL,

    CONSTRAINT patients_pkey PRIMARY KEY (id),
    CONSTRAINT patients_tenant_not_empty CHECK (tenant_id <> ''),
    CONSTRAINT patients_nombre_not_empty CHECK (nombre <> ''),
    CONSTRAINT patients_sexo_valid CHECK (sexo IN ('M', 'F', 'I')),
    CONSTRAINT patients_expediente_unique UNIQUE (tenant_id, num_expediente)
);

CREATE INDEX idx_patients_tenant_id ON patients (tenant_id);
CREATE INDEX idx_patients_nombre    ON patients (tenant_id, nombre);
