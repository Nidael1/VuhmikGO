-- Agrega CURP a patients y users.
-- CURP: Clave Unica de Registro de Poblacion.
-- Identificador unico nacional en Mexico (equivalente a NHI Taiwan).
-- Habilita unicidad de cuenta medica y traspaso de expedientes (ADR-0009).
-- Migracion forward-only. Sin .down.sql.

ALTER TABLE patients ADD COLUMN curp TEXT;
CREATE UNIQUE INDEX idx_patients_curp_tenant ON patients (tenant_id, curp) WHERE curp IS NOT NULL;

ALTER TABLE users ADD COLUMN curp TEXT;
CREATE UNIQUE INDEX idx_users_curp_unique ON users (curp) WHERE curp IS NOT NULL;
