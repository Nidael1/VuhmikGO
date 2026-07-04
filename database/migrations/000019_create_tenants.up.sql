-- ADR-0025: Modelo de datos de Tenant y materialización del Shader Stack.
-- Migración forward-only. Sin .down.sql.

-- 1) Tabla tenants (identidad + shader stack base)
CREATE TABLE tenants (
    tenant_id           TEXT        NOT NULL,
    tenant_area         TEXT        NOT NULL DEFAULT 'medicine',
    country_code        TEXT        NOT NULL DEFAULT 'MX',
    clinical_shader_key TEXT        NOT NULL DEFAULT 'med_basic',
    export_shader_key   TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT tenants_pkey PRIMARY KEY (tenant_id),
    CONSTRAINT tenants_area_check
        CHECK (tenant_area IN ('generic','medicine','nutrition','legal')),
    CONSTRAINT tenants_tenant_id_not_empty CHECK (tenant_id <> '')
);

-- 2) Backfill de tenants existentes (defaults ADR-0025)
INSERT INTO tenants (tenant_id, tenant_area, country_code,
                     clinical_shader_key, export_shader_key)
SELECT DISTINCT tenant_id, 'medicine', 'MX', 'med_basic', 'export_none'
FROM users
ON CONFLICT (tenant_id) DO NOTHING;

-- 3) Tabla tenant_extra_shaders (extra shaders 0..N, fail-closed)
CREATE TABLE tenant_extra_shaders (
    tenant_id  TEXT        NOT NULL REFERENCES tenants(tenant_id),
    shader_key TEXT        NOT NULL,
    active     BOOLEAN     NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, shader_key)
);

CREATE INDEX idx_tenant_extra_shaders_tenant
    ON tenant_extra_shaders (tenant_id);

-- 4) FK users.tenant_id -> tenants.tenant_id (aditiva)
ALTER TABLE users
    ADD CONSTRAINT users_tenant_id_fk
    FOREIGN KEY (tenant_id) REFERENCES tenants(tenant_id);
