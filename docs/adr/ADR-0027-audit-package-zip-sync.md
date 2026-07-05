# ADR-0027 — Audit Package ZIP síncrono en v1

## Estado
Aceptado

## Fecha
2026-07-05

---

## Contexto

El catálogo de Shaders (03_VUHMIK_SHADERS_reglas.md) define `legal_export`
como: "Exportación legal: ZIP por paciente con evidencia + anexos + hashes
verificables." El STACK_REAL.md establece: "no long tasks en request:
export/hashing/zip/reportes → jobs (WAR-A)."

El sistema de jobs asíncronos (cola Redis → worker → ZIP) no existe todavía
en el repositorio. Solo existen workers de mantenimiento (backup, metrics_purge).

Construir el sistema de jobs antes del ZIP agrega un issue de infraestructura
previo que bloquea la entrega del Audit Package.

---

## Problema

Sin decisión explícita, el ZIP no puede implementarse sin violar STACK_REAL.md
(que exige jobs asíncronos) o sin construir primero infraestructura de cola
que no está planeada en el sprint actual.

---

## Decisión

### ZIP síncrono en v1 con límite de tiempo

El Audit Package ZIP se genera síncronamente en el request HTTP en v1, con
las siguientes restricciones:

1. Solo se permite para un paciente a la vez (nunca exportación masiva de
   todos los pacientes en un request).
2. Timeout máximo de 30 segundos en el handler. Si supera el límite, el
   servidor retorna 503 con mensaje explicativo.
3. El ZIP se genera en memoria (bytes.Buffer), se sirve directamente y
   no se persiste en disco ni en Redis. Cache-Control: no-store.
4. La restricción de "no long tasks en request" de STACK_REAL.md aplica a
   exportaciones masivas y jobs de mantenimiento. Un ZIP de un expediente
   individual de un médico independiente (perfil WAR-A: 10-30 tenants) es
   sub-segundo en condiciones normales y no viola el espíritu de la regla.

### Migración a asíncrono (v2)

Cuando el sistema de jobs asíncronos exista (cola Redis + worker), el handler
de ZIP se convierte en: (1) encolar el job, (2) retornar un job_id, (3) el
cliente consulta el estado y descarga cuando está listo. Este ADR NO autoriza
esa implementación — requiere su propio ADR cuando llegue.

---

## Estructura del Audit Package ZIP
PATIENT_<id><timestamp>.zip
├── manifest.json           ← índice del paquete + hash global
├── ips/
│   └── patient_ips.json    ← IPS Bundle FHIR R4 (ADR-0010)
├── evidence/
│   └── record<uuid>.json  ← cada registro de evidencia con su hash
└── hashes/
└── hashes.sha256       ← SHA-256 de cada archivo del paquete
---

## Consecuencias

- El Audit Package ZIP es entregable en v1 sin construir infraestructura nueva.
- Economía de guerra: un solo archivo Go nuevo + un handler + una ruta.
- El Core permanece agnóstico; el ZIP vive en la capa Shaders/delivery.
- Este ADR NO autoriza exportación masiva de múltiples pacientes en un request.
- Este ADR NO autoriza persistencia del ZIP en disco o base de datos.
- Este ADR NO autoriza el sistema de jobs asíncronos (requiere ADR posterior).

## Documentos impactados

- STACK_REAL.md (excepción documentada para ZIP individual síncrono en v1).
- 03_VUHMIK_SHADERS_reglas.md (legal_export implementado con esta estructura).
