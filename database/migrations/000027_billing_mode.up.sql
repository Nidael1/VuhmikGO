-- Migración 000027: billing mode por tenant en tabla users
-- billing_mode: 'monthly' (plan fijo mensual) o 'per_module' (costo por módulo)
-- monthly_fee: precio mensual fijo cuando billing_mode = 'monthly'
-- Si billing_mode = 'per_module', el MRR se calcula sumando tenant_capabilities.costo
-- Si billing_mode = 'monthly', el MRR usa monthly_fee directamente

ALTER TABLE users
  ADD COLUMN IF NOT EXISTS billing_mode TEXT NOT NULL DEFAULT 'per_module'
    CHECK (billing_mode IN ('monthly', 'per_module')),
  ADD COLUMN IF NOT EXISTS monthly_fee  NUMERIC(10,2) NOT NULL DEFAULT 0;

COMMENT ON COLUMN users.billing_mode IS 'Modo de facturación: monthly = plan fijo, per_module = suma de costos por módulo activo';
COMMENT ON COLUMN users.monthly_fee  IS 'Cuota mensual fija (solo aplica cuando billing_mode = monthly)';
