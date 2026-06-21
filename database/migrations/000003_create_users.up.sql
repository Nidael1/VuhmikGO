-- Tabla de usuarios del sistema.
-- Un usuario corresponde a un médico independiente (tenant).
-- La autenticacion es JWT; esta tabla provee la identidad base.
-- Migración forward-only. Sin .down.sql.

CREATE TABLE users (
    id            TEXT        NOT NULL,
    tenant_id     TEXT        NOT NULL,
    email         TEXT        NOT NULL,
    password_hash TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL,

    CONSTRAINT users_pkey PRIMARY KEY (id),
    CONSTRAINT users_email_unique UNIQUE (email),
    CONSTRAINT users_tenant_id_not_empty CHECK (tenant_id <> ''),
    CONSTRAINT users_email_not_empty CHECK (email <> ''),
    CONSTRAINT users_password_not_empty CHECK (password_hash <> '')
);

CREATE INDEX idx_users_tenant_id ON users (tenant_id);
CREATE INDEX idx_users_email     ON users (email);
