# TEST MANUAL CRM — Happy Path

## Versión
1.0 — 2026-06-04

## Propósito
Validar que el flujo CRM administrativo mínimo funciona correctamente
via Shaders, sin lógica de negocio en el frontend y con error_code
visible en casos de error controlado.

No contiene PHI ni datos reales.

---

## Precondiciones

Entorno requerido:
- Go 1.22+ instalado
- PostgreSQL en localhost:5432
- BD: vuhmik_dev (createdb vuhmik_dev)
- Migraciones aplicadas: migrate -path database/migrations up
- Servidor: go run ./cmd/vuhmik-api/
- URL base: http://localhost:8080

Datos dummy (no reales):
- Tenant ID: tenant-test-001
- Actor ID:  actor-test-001
- Subject ID: sujeto-test-001

Headers HTTP requeridos:
  X-Tenant-ID: tenant-test-001
  X-Actor-ID:  actor-test-001

---

## Flujo 1 — Navegación base (GET)

Paso 1: GET /           → HTTP 200, título "Inicio", nav visible
Paso 2: GET /dashboard  → HTTP 200, título "Dashboard"
Paso 3: GET /pacientes  → HTTP 200, título "Pacientes"
Paso 4: GET /ece/nuevo  → HTTP 200, formulario draft, badge BORRADOR visible
Paso 5: GET /ece/emitir → HTTP 200, confirmación de emisión, advertencia visible
Paso 6: GET /ece/anular → HTTP 200, formulario void, selector reason_code visible

---

## Flujo 2 — Creación de registro via Shader (POST con contexto válido)

Comando:
  curl -X POST http://localhost:8080/registros/nuevo
    -H "X-Tenant-ID: tenant-test-001"
    -H "X-Actor-ID: actor-test-001"

Resultado esperado:
  HTTP 202
  Body: "operación autorizada — registro pendiente de persistencia"

---

## Flujo 3 — Captura ECE draft (POST)

Comando:
  curl -X POST http://localhost:8080/ece/nuevo
    -H "X-Tenant-ID: tenant-test-001"
    -H "X-Actor-ID: actor-test-001"
    -d "subject_id=sujeto-test-001&notes=Nota+clinica+de+prueba"

Resultado esperado:
  HTTP 200
  Página de confirmación de borrador registrado

---

## Flujo 4 — Caso de error tipado (sin X-Tenant-ID)

Comando:
  curl -X POST http://localhost:8080/registros/nuevo
    -H "X-Actor-ID: actor-test-001"

Resultado esperado:
  HTTP 403
  error_code visible: ER-SHADER-001
  Mensaje UX: "El contexto de la operación está incompleto o es inválido."
  Sin PHI expuesto en la respuesta

---

## Flujo 5 — Validación UX (campos requeridos vacíos)

Comando:
  curl -X POST http://localhost:8080/ece/nuevo
    -H "X-Tenant-ID: tenant-test-001"
    -H "X-Actor-ID: actor-test-001"
    -d "subject_id=&notes="

Resultado esperado:
  HTTP 422
  Errores de validación UX visibles (subject_id y notes requeridos)
  Sin error_code del Core — es validación UX, no Core

---

## Tabla de resultados

Actualizar con PASS o FAIL al ejecutar.

  Flujo 1 — Navegación base:           [ pendiente ]
  Flujo 2 — Creación via Shader:        [ pendiente ]
  Flujo 3 — Captura ECE draft:          [ pendiente ]
  Flujo 4 — Error tipado ER-SHADER-001: [ pendiente ]
  Flujo 5 — Validación UX:             [ pendiente ]

---

## Notas

- No incluir PHI ni datos reales al actualizar los resultados.
- Los tests automáticos de integración están fuera del alcance de este issue.
- La persistencia real llega en el Sprint de repositorio.
