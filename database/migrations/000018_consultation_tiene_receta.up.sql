-- ADR-0022 / ADR-0024: campo de proyeccion de lectura CQRS.
-- tiene_receta = true si al momento de crear la consulta se emitio una receta vinculada.
-- No es evidencia clinica. No es inmutable por regla del Core.
-- Solo refleja el estado del flujo en ConsultationNewView al momento de guardar.
-- Proyeccion append-only: no se actualiza retroactivamente.

ALTER TABLE consultation_projections
    ADD COLUMN IF NOT EXISTS tiene_receta BOOLEAN NOT NULL DEFAULT FALSE;
