-- Tabla de refresh tokens para sesiones stateful.
-- Permite revocación real por sesión o dispositivo.
-- JWT de acceso: 15 minutos. Refresh: 7 días.
-- Migración forward-only. Sin .down.sql.

CREATE TABLE refresh_tokens (
    id           TEXT        NOT NULL,
    user_id      TEXT        NOT NULL,
    tenant_id    TEXT        NOT NULL,
    token_hash   TEXT        NOT NULL,
    expires_at   TIMESTAMPTZ NOT NULL,
    revoked_at   TIMESTAMPTZ NULL,
    created_at   TIMESTAMPTZ NOT NULL,

    CONSTRAINT refresh_tokens_pkey PRIMARY KEY (id),
    CONSTRAINT refresh_tokens_token_hash_unique UNIQUE (token_hash),
    CONSTRAINT refresh_tokens_user_fk FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_refresh_tokens_user_id    ON refresh_tokens (user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens (token_hash);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens (expires_at);
