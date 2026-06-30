-- ADR-0024: Vincula la receta a la consulta que la originó.
-- Permite al handler de impresión recuperar signos vitales de la consulta.

ALTER TABLE prescription_projections
    ADD COLUMN IF NOT EXISTS consultation_id TEXT;

CREATE INDEX IF NOT EXISTS idx_prescription_proj_consultation
    ON prescription_projections (consultation_id) WHERE consultation_id IS NOT NULL;
