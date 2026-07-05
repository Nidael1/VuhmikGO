# ADR-0026 — Referencia de vendedor en tenant (provisional, fase 1)

## Estado
Aceptado

## Fecha
2026-07-04

---

## Contexto

En la fase 1 de comercialización, el alta de tenants ocurre únicamente por
backend. Cada tenant es captado por un vendedor, y el negocio necesita
registrar qué vendedor originó cada tenant para atribución comercial.

Este concepto es **provisional**: se usará en la fase 1 y se espera un
cambio oficial posterior que lo formalice o reemplace.

---

## Decisión

### 1. Catálogo de vendedores como tabla `vendors`

- `vendor_id` (PK) — formato `vndrNNN` (ej. `vndr001`).
- `name` (TEXT) — nombre del vendedor.
- `active` (BOOLEAN, DEFAULT TRUE).
- `created_at` (TIMESTAMPTZ).

Los vendedores se desactivan, nunca se borran.

### 2. Referencia en `tenants` como columna `vendor_ref` (nullable)

FK → `vendors.vendor_id`. Nullable porque no todo tenant proviene de un
vendedor. Sin vendedor = NULL.

### 3. Carácter provisional y condición de retiro

Cuando llegue el cambio oficial: se evaluará si `vendor_ref` se formaliza,
migra a un modelo comercial más amplio, o se retira. El retiro requiere
su propio ADR.

### 4. Sin lógica en Core ni en Shaders

`vendor_ref` es dato comercial/administrativo. El Core no lo interpreta.
Los Shaders no lo consultan.

---

## Implementación

- Migración `000020_create_vendors.up.sql` (issue #201)
- Seed inicial: `vndr001 / Carlos Ramírez Herrera`

---

## Consecuencias

- Se registra la atribución de vendedor por tenant con catálogo íntegro.
- Este ADR NO autoriza lógica comercial adicional (comisiones, reportes).
- Este ADR NO autoriza que Shaders o Core consulten `vendor_ref`.

---

## Regla final

Este ADR autoriza únicamente el catálogo de vendedores y la columna
`vendor_ref` provisional en `tenants` para atribución comercial de fase 1.
