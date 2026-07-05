# ADR-0025 — Modelo de datos de Tenant y materialización del Shader Stack

## Estado
Aceptado

## Fecha
2026-07-04

---

## Contexto

ADR-0002 (Sistema de Shaders por país y modo genérico) definió que el tenant
**debe** contemplar `country_code` (write-once) y `active_shaders`
(configuración controlada), pero delegó explícitamente la materialización de
esos campos al "issue del plan de ejecución que defina el modelo de tenant",
que a la fecha no existía.

El estado real del repositorio confirmó que ese modelo nunca se materializó:

- No existía tabla `tenants`. El tenant era únicamente un `tenant_id`
  propagado desde `users`.
- Ningún campo del Shader Stack existía en el código Go.
- No existía rama ni issue de catálogo de shaders del Sprint 2.

El canon de Shaders ya define el modelo conceptual completo del Shader Stack:
- `tenant_area` (write-once): profesión/área.
- Exactamente 1 `clinical_shader_key` (obligatorio).
- 0..1 `export_shader_key`.
- 0..N `extra_shader_keys[]`.

---

## Problema

Sin una decisión explícita de materialización:

- No había dónde el tenant declarara su Shader Stack.
- No había dónde vivieran `country_code` ni `tenant_area` como identidad
  write-once del tenant.
- El aislamiento multi-tenant dependía de convención de aplicación, no de
  integridad referencial en base de datos.

---

## Decisión

### 1. El tenant se materializa como entidad de primera clase

Se crea la tabla `tenants` como fuente de verdad de la identidad del tenant
y de su Shader Stack.

### 2. Esquema de la tabla `tenants`

- `tenant_id` (PK)
- `tenant_area` (write-once) — CHECK: `generic`, `medicine`, `nutrition`, `legal`.
- `country_code` (write-once) — identidad de país (ADR-0002 §6).
- `clinical_shader_key` (NOT NULL) — exactamente 1, obligatorio.
- `export_shader_key` (nullable) — 0..1.
- `created_at`, `updated_at` (TIMESTAMPTZ).

### 3. Los extra shaders (0..N) se materializan en tabla separada

Se crea la tabla `tenant_extra_shaders`:
- `tenant_id` (FK → tenants)
- `shader_key` (TEXT NOT NULL)
- `active` (BOOLEAN DEFAULT FALSE) — fail-closed.
- `updated_at` (TIMESTAMPTZ)
- PK compuesta `(tenant_id, shader_key)`

### 4. `users.tenant_id` se convierte en FK hacia `tenants`

Integridad referencial real: `users.tenant_id` referencia `tenants.tenant_id`.

### 5. El Core permanece agnóstico y opaco a los shader keys

Conforme ADR-0002 §2, el Core NO interpreta los valores de shader keys.

### 6. Defaults del backfill de tenants existentes

- `tenant_area = 'medicine'`
- `country_code = 'MX'` (decisión explícita; operación legal solo en México)
- `clinical_shader_key = 'med_basic'`
- `export_shader_key = 'export_none'`
- `tenant_extra_shaders`: vacío (fail-closed)

---

## Implementación

- Migración `000019_create_tenants.up.sql` (issue #200)
- Migración `000021_seed_mx_medical_tenants.up.sql` (issue #207)
- Puerto `ports.TenantRepository` (issue #204)
- Adaptador `postgres.TenantRepository` (issue #204)

---

## Consecuencias

- El tenant pasa a ser entidad de primera clase con integridad referencial.
- El Shader Stack tiene soporte de datos y puede conmutarse por tenant.
- El Core permanece agnóstico.
- Este ADR NO autoriza triggers SQL de write-once enforcement.
- Este ADR NO autoriza modificar el lifecycle ni el modelo append-only.

---

## Documentos impactados

- `ADR-0002-shaders-por-pais-y-modo-generico.md` (materializa su decisión §6-7).
- `03_VUHMIK_SHADERS_reglas.md` (el Shader Stack canónico obtiene soporte de datos).

---

## Regla final

Este ADR autoriza únicamente el diseño del modelo de datos de tenant y la
materialización del Shader Stack descritos. No autoriza cambios al
agnosticismo del Core, al lifecycle, ni al modelo append-only.
