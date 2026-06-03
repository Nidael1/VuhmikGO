# ADR-0001 — Stack canónico Go para VUHMÍK

## Estado
Aceptado

## Fecha
2026-06-02

---

## Contexto

La documentación previa de VUHMÍK contenía decisiones técnicas orientadas a Python/FastAPI para mejorar el ambiente inicial de ejecución.

La decisión vigente cambia el stack de implementación a Go, manteniendo intactos los principios no negociables del producto:

- VUHMÍK es un solo producto.
- La ejecución se organiza en tres capas/fases: ENGINE, Shaders y Asteroids.
- Core agnóstico, determinista e inmutable.
- Acceso al Core únicamente vía Shaders.
- Asteroids como UX/producto, sin reglas clínicas ni legales.
- Evidencia append-only.
- Export legal bajo demanda y sin persistencia.
- Observabilidad sin PHI/PII.
- Multi-tenant fail-closed.
- WAR-A con Redis y workers obligatorios.

---

## Problema

Mantener referencias a Python/FastAPI en los documentos de ejecución generaría ambigüedad para una IA ejecutora y riesgo de implementación contradictoria.

VUHMÍK requiere un paquete documental limpio, consistente y ejecutable en Go, sin cambiar el alcance funcional ni las reglas clínicas, legales, operativas o probatorias.

---

## Decisión

Se adopta Go como stack canónico de implementación para VUHMÍK.

Stack aprobado:

- Lenguaje: Go 1.22+ o versión estable definida en `go.mod`.
- HTTP boundary: Go `net/http` + `ServeMux`.
- Contratos: DTOs tipados en Go.
- Validación estructural: validadores explícitos + JSON Schema para payloads agnósticos.
- PostgreSQL: `pgx/v5` con SQL explícito.
- Migraciones: `golang-migrate/migrate` con migraciones SQL forward-only.
- Jobs: workers Go con cola Redis.
- Storage: `StoragePort` como interfaz obligatoria.
- Observabilidad: logs JSON estructurados a stdout, métricas agregadas en Postgres.
- Empaquetado: Docker.
- Operación: Coolify + Hetzner + backups externos cifrados, conforme WAR-A.

---

## Alternativas consideradas

### Mantener Python/FastAPI

Rechazado para esta nueva base de ejecución porque el objetivo actual es desarrollar VUHMÍK en Go.

### Usar framework HTTP pesado en Go

Rechazado por economía de guerra, simplicidad operativa y reducción de dependencias.

### Usar ORM

Rechazado para el Core porque las queries críticas, RLS, transacciones e invariantes deben ser explícitas y auditables.

---

## Consecuencias

- Toda referencia a Python, FastAPI, Pydantic, SQLAlchemy, Alembic, Celery o patrones propios de Python debe eliminarse de los documentos operativos de ejecución.
- El cambio de stack no autoriza nuevos features.
- El cambio de stack no autoriza modificar lifecycle, tenancy, append-only, export legal ni observabilidad.
- Las tareas existentes deben conservar su micro-sprint y criterio QA, adaptando únicamente la implementación técnica a Go.
- Claude u otra IA ejecutora debe recibir solo la documentación limpia de Go y ejecutar issue por issue.

---

## Documentos impactados

- `01_VUHMIK_quienes_somos.md.md`
- `02_ENGINE_CORE_que_hace_cerebro.md`
- `11_seguridad.md`
- Cualquier guía futura de ejecución, stack, infraestructura o prompt maestro.

---

## Regla final

Este ADR autoriza ún
icamente el cambio de stack a Go.

No autoriza cambios de alcance, arquitectura por capas, reglas clínicas, reglas legales, lifecycle, modelo de evidencia, WAR-A ni disciplina de ejecución.
