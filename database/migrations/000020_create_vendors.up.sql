-- ADR-0026: Referencia de vendedor en tenant (provisional fase 1).
-- Migración forward-only. Sin .down.sql.

-- 1) Catálogo de vendedores (alta dinámica)
CREATE TABLE vendors (
    vendor_id  TEXT        NOT NULL,
    name       TEXT        NOT NULL DEFAULT '',
    active     BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT vendors_pkey PRIMARY KEY (vendor_id),
    CONSTRAINT vendors_vendor_id_not_empty CHECK (vendor_id <> '')
);

-- 2) Seed inicial (un vendedor en fase 1)
INSERT INTO vendors (vendor_id, name) VALUES
    ('vndr001', 'Carlos Ramírez Herrera')
ON CONFLICT (vendor_id) DO NOTHING;

-- 3) Referencia en tenants (nullable, FK a vendors)
ALTER TABLE tenants
    ADD COLUMN vendor_ref TEXT;

ALTER TABLE tenants
    ADD CONSTRAINT tenants_vendor_ref_fk
    FOREIGN KEY (vendor_ref) REFERENCES vendors(vendor_id);
