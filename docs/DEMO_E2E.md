# DEMO E2E — Flujo completo reproducible

## Fecha
2026-06-04

## Propósito

Demostrar el flujo completo del MVP: captura, emisión, bloqueo,
anulación, reemplazo y export legal — de inicio a fin, con datos
ficticios y comportamiento determinista.

No usa datos reales. No persiste exports.

## Datos ficticios

  Tenant ID:  tenant-demo-001
  Actor ID:   actor-demo-001
  Subject ID: paciente-demo-001

## Precondiciones

  export DATABASE_URL="postgres://localhost:5432/vuhmik_dev?sslmode=disable"
  migrate -path database/migrations -database "$DATABASE_URL" up
  go run ./cmd/vuhmik-api/

## Paso 1 — Captura draft

  curl -X POST http://localhost:8080/ece/draft/guardar \
    -H "X-Tenant-ID: tenant-demo-001" \
    -H "X-Actor-ID: actor-demo-001" \
    -d "subject_id=paciente-demo-001&notes=Nota+clinica+ficticia"

  Esperado: HTTP 201
  {"id":"draft-paciente-demo-001","state":"draft","message":"borrador creado — persistencia pendiente de repositorio"}

## Paso 2 — Confirmación de emisión (UI)

  curl http://localhost:8080/ece/emitir \
    -H "X-Tenant-ID: tenant-demo-001" \
    -H "X-Actor-ID: actor-demo-001"

  Esperado: HTTP 200, pantalla de confirmación con advertencia
  de irreversibilidad.

## Paso 3 — Emisión y bloqueo

  curl -X POST http://localhost:8080/ece/emitir \
    -H "X-Tenant-ID: tenant-demo-001" \
    -H "X-Actor-ID: actor-demo-001"

  Esperado: HTTP 200, "Nota emitida y bloqueada"

  Resultado de dominio: draft → issued → locked (inmutable)

## Paso 4 — Intento de edición post-bloqueo (debe fallar)

  Cualquier intento de Update sobre un registro locked retorna
  ER-CORE-001 (GuardMutation), por diseño.

## Paso 5 — Anulación y reemplazo

  curl -X POST http://localhost:8080/ece/anular \
    -H "X-Tenant-ID: tenant-demo-001" \
    -H "X-Actor-ID: actor-demo-001" \
    -d "reason_code=RC-VOID-002&replacement_notes=Nota+actualizada+ficticia"

  Esperado: HTTP 200
  "Nota anulada — reemplazo emitido"

  Resultado de dominio:
    - Original: voided, voided_at asignado, replaced_by_id enlazado
    - Reemplazo: nuevo registro issued
    - Historial completo preservado (append-only)

## Paso 6 — Export legal (efímero)

  curl -X POST http://localhost:8080/ece/exportar \
    -H "X-Tenant-ID: tenant-demo-001" \
    -H "X-Actor-ID: actor-demo-001" \
    -d "evidence_id=draft-paciente-demo-001"

  Esperado: HTTP 200, Content-Disposition: attachment,
  Cache-Control: no-store

  El archivo JSON se genera en memoria y se descarta tras la
  respuesta. No queda ningún archivo en disco.

## Determinismo

Repetir los pasos 1-6 con los mismos datos ficticios produce
exactamente el mismo comportamiento: mismas transiciones de estado,
mismos error_code en casos de error, mismo formato de export.

## Resultado de la demo

| Paso | Estado |
|------|--------|
| 1 — Captura draft | Reproducible |
| 2 — Confirmación emisión | Reproducible |
| 3 — Emisión + lock | Reproducible |
| 4 — Inmutabilidad post-lock | Reproducible (ER-CORE-001) |
| 5 — Void + replace | Reproducible |
| 6 — Export efímero | Reproducible, sin persistencia |
