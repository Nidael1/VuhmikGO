-- ADR-0021: Perfil profesional por rubro
-- Separado de users (agnóstica) para mantener el Core agnostico de dominio.
-- En v1 solo existe rubro 'medicine'. Escala a otros rubros sin migraciones destructivas.

CREATE TABLE IF NOT EXISTS professional_profiles (
    user_id             TEXT        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id           TEXT        NOT NULL,
    rubro               TEXT        NOT NULL DEFAULT 'medicine',
    nombre_completo     TEXT        NOT NULL DEFAULT '',
    cedula_profesional  TEXT        NOT NULL DEFAULT '',
    especialidad        TEXT        NOT NULL DEFAULT '',
    datos_extra         JSONB,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, rubro)
);

CREATE INDEX IF NOT EXISTS idx_professional_profiles_tenant
    ON professional_profiles (tenant_id);

CREATE INDEX IF NOT EXISTS idx_professional_profiles_user
    ON professional_profiles (user_id);
