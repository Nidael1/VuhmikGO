-- Agrega campos de contenido clinico a la tabla evidence.
-- subject_id: identificador del paciente (no nombre real, solo ID)
-- notes: contenido libre de la nota clinica
-- Migracion forward-only. Sin .down.sql.

ALTER TABLE evidence
  ADD COLUMN subject_id TEXT NOT NULL DEFAULT '',
  ADD COLUMN notes      TEXT NOT NULL DEFAULT '';
