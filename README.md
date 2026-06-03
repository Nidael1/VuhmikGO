# VuhmikGO

Repositorio oficial de **VUHMÍK** — sistema HealthTech para médicos
independientes, distribuido como SaaS y white label sobre un único
binario Go.

> **Estado del repositorio:** en construcción por sprints documentados.
> Este README no describe features. La verdad operativa vive en la
> documentación canónica del proyecto (ADRs y plan de ejecución).

---

## Documentación canónica

- `docs/adr/ADR-0001-stack-go.md` — Stack canónico Go.
- `docs/adr/ADR-0002-shaders-por-pais-y-modo-generico.md` — Sistema de
  Shaders por país y modo genérico.
- `docs/adr/ADR-0003-estructura-go-idiomatica.md` — Estructura de
  carpetas Go idiomática y topología monolito modular hexagonal.

---

## Estructura del repositorio

```
VuhmikGO/
├── cmd/                 → entrypoints ejecutables (futuro)
├── internal/
│   ├── core/            → reglas puras, entidades, lifecycle (Core)
│   ├── application/     → casos de uso y puertos (futuro)
│   ├── infrastructure/  → adaptadores concretos (futuro)
│   └── delivery/        → entrada HTTP (futuro)
├── docs/
│   └── adr/             → Architecture Decision Records
├── go.mod
└── README.md
```

---

## Reglas operativas

- Un issue = un contrato = una rama = un PR = un commit (salvo excepción
  documentada en el issue).
- El Core es agnóstico, determinista, inmutable y append-only.
- Acceso al Core únicamente vía Shaders.
- Asteroides son capa UX/producto, sin reglas clínicas ni legales.
- Multi-tenant estricto, fail-closed.
- Observabilidad sin PHI/PII.

---

## Módulo

```
module github.com/Nidael1/VuhmikGO
```
