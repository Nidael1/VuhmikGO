-- ADR-0021: Campos adicionales del perfil profesional
-- Necesarios para el PDF de receta (NOM-024) y el registro de usuarios por admin.

ALTER TABLE professional_profiles
    ADD COLUMN IF NOT EXISTS universidad TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS direccion   TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS telefono    TEXT NOT NULL DEFAULT '';
