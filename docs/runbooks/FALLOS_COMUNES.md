# RUNBOOK — Fallos comunes

## 1. ER-SHADER-001 — contexto incompleto

Síntoma: HTTP 403 con error_code ER-SHADER-001.

Causa: faltan headers X-Tenant-ID o X-Actor-ID en la solicitud.

Acción: verificar que el cliente envía ambos headers en toda
operación sensible. Las rutas /, /dashboard y /pacientes no
requieren estos headers.

## 2. ER-CORE-001 — registro inmutable

Síntoma: operación rechazada con ER-CORE-001.

Causa: se intentó modificar un registro en estado issued o locked.

Acción: esto es correcto por diseño. Para corregir, usar el flujo
de void + replace (/ece/anular) con reason_code obligatorio.

## 3. ER-CORE-002 — transición inválida

Síntoma: operación rechazada con ER-CORE-002.

Causa: se intentó una transición de estado no declarada en la
matriz de lifecycle (ej. voided → issued).

Acción: revisar internal/core/evidence/lifecycle.go para conocer
las transiciones válidas. No existe transición implícita.

## 4. ER-CORE-003 — reason_code ausente o inválido

Síntoma: operación void/replace rechazada con ER-CORE-003.

Causa: el reason_code enviado no está en el catálogo
ReasonCodeCatalog (internal/core/evidence/reason.go).

Acción: usar uno de los códigos RC-VOID-* o RC-REPLACE-* definidos
en el catálogo.

## 5. Fallo de conexión a PostgreSQL

Síntoma: error al arrancar o al ejecutar operaciones de repositorio.

Acción:
  1. Verificar que PostgreSQL está corriendo: pg_isready
  2. Verificar DATABASE_URL apunta a la base correcta
  3. Verificar que las migraciones están aplicadas:
     migrate -path database/migrations -database "$DATABASE_URL" version
