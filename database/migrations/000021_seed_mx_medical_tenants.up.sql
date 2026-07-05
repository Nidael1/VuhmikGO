-- ADR-0025 + ADR-0002: Activar mx_medical y legal_export en tenants existentes.
-- Fase 1 México — todos los tenants operan bajo cumplimiento NOM-024-SSA3-2012.
-- Migración forward-only. Sin .down.sql.

-- 1) Activar mx_medical como extra shader para todos los tenants existentes.
-- mx_medical es el shader de cumplimiento normativo México (ADR-0002 §4).
INSERT INTO tenant_extra_shaders (tenant_id, shader_key, active, updated_at)
SELECT tenant_id, 'mx_medical', TRUE, NOW()
FROM tenants
ON CONFLICT (tenant_id, shader_key) DO UPDATE
    SET active = TRUE, updated_at = NOW();

-- 2) Actualizar export_shader_key de export_none a legal_export.
-- Los tenants médicos MX tienen derecho a export legal (ADR-0007).
UPDATE tenants
SET export_shader_key = 'legal_export', updated_at = NOW()
WHERE export_shader_key = 'export_none';
