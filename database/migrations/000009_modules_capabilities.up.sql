-- ADR-0017: Registro de capacidades por tenant
-- Plano de control: catalogo global de modulos (solo lectura en app)
-- Plano de datos: activacion por tenant (escrito por admin, ADR-0018)

-- Tabla de modulos publicados (plano de control)
CREATE TABLE IF NOT EXISTS modules (
    id                 TEXT PRIMARY KEY,
    rubro              TEXT NOT NULL DEFAULT 'medico',
    publication_status TEXT NOT NULL DEFAULT 'en_desarrollo',
    descripcion        TEXT NOT NULL DEFAULT '',
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT modules_status_check
        CHECK (publication_status IN ('en_desarrollo','publicado','deprecado'))
);

-- Tabla de activacion por tenant (plano de datos)
-- fail-closed: active = false por defecto
CREATE TABLE IF NOT EXISTS tenant_capabilities (
    tenant_id  TEXT NOT NULL,
    module_id  TEXT NOT NULL REFERENCES modules(id),
    active     BOOLEAN NOT NULL DEFAULT FALSE,
    plan       TEXT NOT NULL DEFAULT '',
    costo      NUMERIC(10,2) NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, module_id)
);

CREATE INDEX IF NOT EXISTS idx_tenant_capabilities_tenant
    ON tenant_capabilities (tenant_id);

-- Seed: modulos del rubro medico publicados (ADR-0020)
-- El super-admin en v1 = esta migracion
INSERT INTO modules (id, rubro, publication_status, descripcion) VALUES
    ('note',          'medico', 'publicado', 'Notas clinicas'),
    ('prescription',  'medico', 'publicado', 'Receta electronica'),
    ('allergy',       'medico', 'publicado', 'Alergias e intolerancias'),
    ('diagnosis',     'medico', 'publicado', 'Diagnosticos y lista de problemas'),
    ('immunization',  'medico', 'publicado', 'Inmunizaciones y vacunacion'),
    ('lab_result',    'medico', 'publicado', 'Resultados de laboratorio')
ON CONFLICT (id) DO NOTHING;
