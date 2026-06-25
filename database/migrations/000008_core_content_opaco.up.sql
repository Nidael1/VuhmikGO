-- ADR-0016: Core agnostico — contenido opaco
-- Paso 1: agregar columna content (nullable inicialmente)
ALTER TABLE evidence ADD COLUMN IF NOT EXISTS content JSONB;

-- Paso 2: poblar content envolviendo notes existentes
UPDATE evidence
SET content = jsonb_build_object('type', 'note', 'text', COALESCE(notes, ''))
WHERE content IS NULL;

-- Paso 3: volver content NOT NULL
ALTER TABLE evidence ALTER COLUMN content SET NOT NULL;

-- Paso 4: renombrar subject_id a subject_ref
ALTER TABLE evidence RENAME COLUMN subject_id TO subject_ref;

-- Paso 5: drop notes (ya migrado a content)
ALTER TABLE evidence DROP COLUMN IF EXISTS notes;

-- Indice sobre subject_ref para busquedas por sujeto
CREATE INDEX IF NOT EXISTS idx_evidence_subject_ref
  ON evidence (tenant_id, subject_ref);
