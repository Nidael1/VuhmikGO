-- Core: índices críticos de evidencia
-- Migración: solo hacia adelante — sin rollback en producción
-- Índices obligatorios: tenant_id, state, issued_at
-- ADR-0001: PostgreSQL, migraciones forward-only

CREATE INDEX IF NOT EXISTS idx_evidence_tenant_id
    ON evidence (tenant_id);

CREATE INDEX IF NOT EXISTS idx_evidence_state
    ON evidence (state);

CREATE INDEX IF NOT EXISTS idx_evidence_issued_at
    ON evidence (issued_at);
