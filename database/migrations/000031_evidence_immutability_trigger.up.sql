-- Trigger de inmutabilidad para evidencia emitida.
-- Protege registros en estado issued, locked y voided contra
-- UPDATE y DELETE, independientemente de quien ejecute el query.
-- Esto refuerza el principio append-only del Core a nivel de base
-- de datos, no solo a nivel de aplicacion.

CREATE OR REPLACE FUNCTION enforce_evidence_immutability()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.state IN ('issued', 'locked', 'voided') THEN
        RAISE EXCEPTION 'EVIDENCE_IMMUTABLE: el registro % en estado % no puede modificarse ni eliminarse',
            OLD.id, OLD.state;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER evidence_immutability_before_update
    BEFORE UPDATE ON evidence
    FOR EACH ROW
    EXECUTE FUNCTION enforce_evidence_immutability();

CREATE TRIGGER evidence_immutability_before_delete
    BEFORE DELETE ON evidence
    FOR EACH ROW
    EXECUTE FUNCTION enforce_evidence_immutability();
