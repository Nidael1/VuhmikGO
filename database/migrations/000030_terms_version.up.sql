-- terms_version identifica pais y rubro del documento aceptado
-- (ej. mx-medicine-v1). Permite invalidar aceptaciones previas cuando
-- se publica una version nueva, y distinguir contextos legales distintos
-- cuando existan tenants fuera de Mexico.
-- Pendiente: resolucion por shader consent_policy_* (requiere ADR).
ALTER TABLE users ADD COLUMN IF NOT EXISTS terms_version TEXT;
