-- ADR-0018: Panel de toggles — flags de administración en users
-- is_admin: acceso al panel de control comercial
-- is_suspended: bloquea login sin borrar datos (falta de pago, etc.)

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS is_admin     BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS is_suspended BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_users_is_admin
    ON users (is_admin) WHERE is_admin = TRUE;

-- El usuario de desarrollo tiene acceso de admin
UPDATE users SET is_admin = TRUE
    WHERE email = 'dev@vuhmik.com';
