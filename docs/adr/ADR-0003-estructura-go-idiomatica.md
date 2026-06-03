# ADR-0003 — Estructura de carpetas Go idiomática para VuhmikGO

## Estado
Aceptado

## Fecha
2026-06-02

---

## Contexto

ADR-0001 adoptó Go como stack canónico de VUHMÍK y ordenó eliminar de los
documentos operativos toda referencia a patrones no-Go (Python, FastAPI,
Pydantic, SQLAlchemy, Alembic, Celery).

La documentación heredada (`02_ENGINE_CORE`, §16 `PROJECT_STRUCTURE`)
describe un layout de carpetas con convenciones no propias de Go:

```
/engine-core
 └─ /app
     ├─ /Domain
     │   ├─ Document
     │   ├─ Tenant
     │   ├─ ValueObjects
     │   └─ Enums
     ├─ /Application
     ├─ /Infrastructure
     └─ /Http
```

Este layout (PascalCase, agrupación `/app/Domain/...`) corresponde a stacks
PHP/Laravel/Symfony, no a Go.

En el mismo documento (`02_ENGINE_CORE`, §6 `Convenciones por Carpeta`) ya
aparece una convención Go idiomática:

```
internal/engine/records/
internal/engine/schema_runtime/
internal/engine/legal_context/
internal/engine/tenancy/
cmd/engine-api/main.go
```

Las dos estructuras coexisten en el canon. Esto produce ambigüedad al
ejecutar Issue #1 y siguientes, porque el plan de ejecución
(`12_execution.md`) usa un placeholder `<core_models_path>` sin resolver.

---

## Problema

Sin una decisión explícita:

- No hay una única ruta canónica para crear las entidades del Core.
- Issue #1 (`Definir entidades Core base`) no puede ejecutarse sin
  inventar arquitectura.
- Existe riesgo de mezclar dos layouts incompatibles en el mismo
  repositorio.

---

## Decisión

### 1. Estructura Go idiomática como canónica para VuhmikGO

Se adopta como **canónica** la estructura Go idiomática basada en `/cmd`
e `/internal`, alineada con ADR-0001 y con la §6 de `02_ENGINE_CORE`.

### 2. Estructura PascalCase queda obsoleta para VuhmikGO

El layout `/engine-core/app/Domain | Application | Infrastructure | Http`
queda **obsoleto** para VuhmikGO por ser herencia de documentación
anterior al cambio de stack.

Esta obsolescencia aplica únicamente al repositorio `VuhmikGO`. La
documentación histórica se conserva como contexto, sin valor ejecutorio.

### 3. Arquitectura hexagonal preservada, traducida a Go idiomático

La arquitectura hexagonal se mantiene en su **semántica** y se expresa con
nombres Go idiomáticos:

| Concepto hexagonal | Carpeta Go canónica          | Contenido                                                                 |
| ------------------ | ---------------------------- | ------------------------------------------------------------------------- |
| Domain             | `internal/core/...`          | Reglas puras, entidades, value objects, lifecycle                         |
| Application        | `internal/application/...`   | Casos de uso y puertos (interfaces)                                       |
| Infrastructure     | `internal/infrastructure/...`| Adaptadores concretos: PostgreSQL (pgx), Redis, StoragePort               |
| Delivery (HTTP)    | `internal/delivery/http/...` | Entrada HTTP con `net/http` + `ServeMux`                                  |
| Entrypoints        | `cmd/...`                    | Binarios ejecutables (`main.go`)                                          |

Esquema esperado del repositorio (no exhaustivo, se construye por issues):

```
VuhmikGO/
├── cmd/
│   └── vuhmik-api/
│       └── main.go
├── internal/
│   ├── core/
│   │   └── evidence/
│   │       ├── entity.go
│   │       └── state.go
│   ├── application/
│   ├── infrastructure/
│   └── delivery/
│       └── http/
├── docs/
│   └── adr/
├── go.mod
└── go.sum
```

### 4. Ruta autorizada para Issue #1

Para Issue #1 (`Definir entidades Core base`) el path autorizado es:

```
internal/core/evidence/entity.go
```

### 5. Archivo de tipo de estado permitido en Issue #1

Si Issue #1 requiere un tipo para que el campo `state` compile en Go, se
permite crear:

```
internal/core/evidence/state.go
```

únicamente con **lo mínimo necesario** para tipar el campo `state`. NO se
permite en Issue #1:

- Definir valores/constantes de estado (eso es Issue #2).
- Definir lifecycle.
- Definir matriz de transiciones.
- Agregar comportamiento.

### 6. Issue #1 no autoriza nada más

Issue #1 NO autoriza:

- Crear modelo de tenant ni configuración.
- Crear `active_shaders` ni `country_code`.
- Crear API, migraciones, Shaders, Asteroides.
- Agregar métodos ni comportamiento a la entidad.

### 7. Módulo Go canónico

```
module github.com/Nidael1/VuhmikGO
```

### 8. Topología técnica autorizada: Monolito Modular Hexagonal

Se adopta **Monolito Modular Hexagonal** en Go como topología técnica:

- Un solo repositorio: `github.com/Nidael1/VuhmikGO`.
- Un solo binario principal inicialmente.
- Separación interna por `/cmd` e `/internal`.
- Core/domain puro, sin dependencias externas.
- Application: casos de uso + puertos.
- Infrastructure: adaptadores concretos.
- Delivery/HTTP: única entrada del sistema.
- Shaders: frontera contractual entre Core y políticas.
- Asteroides: capa UX/producto.

### 9. Topologías y prácticas NO autorizadas

Este ADR **rechaza explícitamente** para la fase actual:

- Microservicios.
- Kubernetes / orquestación compleja.
- Autoescalado de infraestructura.
- Service mesh.
- CQRS (Command-Query Responsibility Segregation).
- Event sourcing complejo.
- Múltiples repositorios.

**La escalabilidad debe venir de límites claros entre capas, no de
infraestructura compleja desde el día uno.**

Cualquier introducción futura de estos patrones requiere un ADR posterior
con justificación específica.

---

## Alternativas consideradas

### Mantener el layout `/app/Domain | Application | Infrastructure | Http`
Rechazado. Es herencia de stack no-Go y contradice ADR-0001.

### Coexistencia de ambos layouts en el mismo repo
Rechazado. Produce ambigüedad, viola el principio de un solo orden y
contradice la disciplina de freeze.

### Microservicios desde el inicio
Rechazado. Economía de guerra para un solo desarrollador, sin justificación
funcional y con costo operativo desproporcionado.

### Capas más finas (ej. CQRS, event sourcing)
Rechazado. Sobre-ingeniería sin justificación funcional para el alcance
actual del producto.

---

## Consecuencias

- La §16 `PROJECT_STRUCTURE` de `02_ENGINE_CORE` queda **obsoleta para
  VuhmikGO**. La §6 `Convenciones por Carpeta` del mismo documento es
  consistente con este ADR y se mantiene como referencia.
- Todas las referencias a `<core_models_path>` en `12_execution.md`
  resuelven a `internal/core/...` según el dominio del issue.
- El repositorio se inicializa con `go.mod` y `module
  github.com/Nidael1/VuhmikGO` como prerequisito de cualquier issue de
  código.
- La arquitectura hexagonal se preserva en semántica, no en estética de
  carpetas.
- La escalabilidad del sistema depende de mantener límites claros entre
  `core`, `application`, `infrastructure` y `delivery`, no de añadir
  infraestructura.

---

## Documentos impactados

- `02_ENGINE_CORE_que_hace_cerebro.md` (§16 obsoleta para VuhmikGO; §6
  consistente con este ADR).
- `12_execution.md` (placeholders `<core_models_path>` se resuelven contra
  esta estructura).

---

## Regla final

Este ADR autoriza únicamente la estructura de carpetas Go idiomática y la
topología de monolito modular hexagonal descritas.

No autoriza cambios al agnosticismo del Core, al lifecycle, al modelo de
evidencia append-only, al alcance de issues ya definidos, ni la
introducción de microservicios, autoescalado, event sourcing complejo,
CQRS ni múltiples repositorios.
