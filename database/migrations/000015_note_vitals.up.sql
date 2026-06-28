-- Agregar signos vitales a note_projections
-- Los signos vitales son parte de la nota clínica (exploración física)
-- y se proyectan para el PDF de receta y el IPS.

ALTER TABLE note_projections
    ADD COLUMN IF NOT EXISTS ta    TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS fc    TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS fr    TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS temp  TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS peso  TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS talla TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS sao2  TEXT NOT NULL DEFAULT '';

-- Agregar referencia opcional de receta a nota clínica
ALTER TABLE prescription_projections
    ADD COLUMN IF NOT EXISTS clinical_note_id TEXT;
