# CLAUDE.md — VUHMÍK (VuhmikGO)

Este archivo es la fuente de verdad para Claude Code al trabajar en este repositorio.
Lee este archivo completo antes de cualquier acción.

---

## Qué es este proyecto

VUHMÍK es un sistema SaaS de Expediente Clínico Electrónico (ECE) para médicos
independientes en México. Cumple NOM-024-SSA3-2012 y NOM-004-SSA3-2012.

Es un solo producto. No es un MVP. No son módulos separados.
La organización interna en capas es arquitectónica, no comercial.

Empresa: NDT (Next Dev Tech). Fundadora y desarrolladora única: Nida.

---

## Arquitectura — tres capas concéntricas (NO negociable)

```
Asteroides (UX/frontend)
      ↓ solo via HTTP → Shaders
Shaders (políticas, contratos, reglas clínicas/legales)
      ↓ única vía permitida
Core (internal/core/evidence/) — motor agnóstico, append-only, inmutable
```

### Core (`internal/core/evidence/`)
- Motor de eventos clínicos genérico.
- Agnóstico: no conoce países, profesiones, reglas clínicas ni legales.
- Append-only: nunca UPDATE ni DELETE en evidencia emitida.
- Determinista e inmutable.
- Solo sabe de "evidencias" (eventos médicos genéricos).

### Shaders (`internal/shaders/`)
- Única vía permitida hacia el Core.
- Dan semántica clínica al Core según país y contexto.
- Shaders activos: `med_basic`, `legal_export`, `export_none`, `default_generic_clinical`.
- Un Shader PUEDE endurecer lo que el Core permite. NUNCA puede permitir lo que el Core prohíbe.
- No contienen UI. No mutan estado. No acceden a infraestructura directamente.

### Asteroides (frontend/UX)
- Módulos de producto activables por tenant: `crm_ui`, `scheduler_ui`, `attachments_ui`, `legal_ui`, `bi_ui`, `billing_ui`.
- Nunca acceden al Core directamente.
- Nunca contienen lógica clínica ni reglas de negocio.
- Solo validan experiencia de usuario (required, formato visual).

---

## Stack oficial

- **Backend**: Go 1.22+, `net/http` + `ServeMux`, binario único.
- **Base de datos**: PostgreSQL con `pgx/v5`. Cast `::text` para `date`, `::float8` para `numeric`.
- **Migraciones**: `golang-migrate`, forward-only. Sin `.down.sql`.
- **Cache/Tokens**: Redis (obligatorio en producción — WAR-A).
- **Workers**: Go workers con cola Redis (obligatorio).
- **Frontend**: Vue 3 + TypeScript + Vite. NVM en `/Volumes/D/nvm`. Node 22.
- **Logs**: JSON estructurados a stdout. Sin PHI/PII en logs.
- **Métricas**: agregadas en PostgreSQL.
- **Infra**: Docker, Hetzner CX23, Coolify 4.x, Traefik, Cloudflare DNS.

### Prohibido en este proyecto
- Python, FastAPI, Pydantic, SQLAlchemy, Alembic, Celery.
- UPDATE o DELETE en tablas de evidencia emitida.
- Migraciones `.down.sql`.
- Logs con contenido clínico o datos personales.
- Exports legales persistidos en disco.
- Acceso directo al Core desde Asteroides.
- Lógica clínica en el frontend.

---

## Lifecycle de evidencia (invariante absoluta)

```
draft → issued → locked / voided
```

- `draft`: borrador, mutable.
- `issued`: emitido, inmutable. No se edita ni se borra.
- `locked`: bloqueado (responsabilidad del backend).
- `voided`: anulado. Genera nuevo evento; el original permanece.
- Correcciones: void + replace (nueva evidencia). Nunca edición directa.
- Export legal: bajo demanda, sin persistencia del archivo generado.

Campos obligatorios en toda evidencia:
`id, tenant_id, state, created_at, issued_at, voided_at, replaced_by_id`

---

## Multi-tenancy

- Cada médico = un tenant aislado.
- `tenant_id` presente en toda operación.
- Fail-closed: si la verificación de tenant falla, se rechaza la operación.
- RLS habilitado en PostgreSQL.
- `tenant_area` es write-once al crear el tenant.

---

## WAR-A — Perfil de operación único

VUHMÍK opera exclusivamente bajo WAR-A. No existen perfiles alternativos.

WAR-A requiere obligatoriamente:
- Redis activo.
- Workers en segundo plano activos.
- Automatismos de integridad.
- Backups automáticos.
- Purge de métricas.
- Revocación de accesos.

Operar sin Redis o sin workers = operación inválida.

---

## Disciplina de ejecución

```
1 issue = 1 rama = 1 PR = 1 merge = 1 cierre = 1 commit (salvo indicación explícita)
```

### Naming de ramas
```
git checkout -b issue/NNN-nombre-corto
```

### Formato de commits
- Código: `[sprint X.Y][TX.Y] descripción (issue #NNN)`
- Docs:   `[docs][issue #NNN] descripción corta`

### Pull Requests
- El PR description DEBE iniciar con: `Closes #NNN`
- Un PR cierra exactamente un issue.

### Flujo Git
```bash
git checkout main && git pull
git checkout -b issue/NNN-nombre
# implementar
git add .
git commit -m "[sprint X.Y][TX.Y] descripción (issue #NNN)"
git push origin issue/NNN-nombre
# abrir PR via GitHub API con $GH_TOKEN
git checkout main && git merge --no-ff issue/NNN-nombre
git push origin main
```

### Edición de archivos
- Python `open()` con rutas absolutas.
- `assert old in content` antes de cada escritura.
- No usar heredocs (poco confiables en esta terminal).

### Reinicio de backend
```bash
lsof -ti:8080 | xargs kill -9
```

### Frontend (dentro de `frontend/`)
- `git add` con rutas relativas (sin prefijo `frontend/`).

---

## Reglas regulatorias

- **NOM-024-SSA3-2012**: estándar de información clínica. Las validaciones deben vivir en el Shader `med_basic`, no en handlers.
- **NOM-004-SSA3-2012**: norma de expediente clínico.
- **LFPDPPP**: protección de datos personales.
- No reclamar certificación DGIS hasta obtenerla formalmente.
- No reclamar soporte de decisión clínica.
- No exponer PHI/PII en logs, métricas ni exports no autorizados.

---

## Observabilidad

- Logs JSON estructurados. Nunca incluir texto clínico.
- Métricas agregadas (sin PHI).
- Todo error tiene `error_code`.
- Toda acción sensible tiene `reason_code`.
- Eventos de auditoría mínimos en todas las operaciones sensibles.

---

## ADRs y documentación

- Todo cambio estructural, de esquema o arquitectónico requiere un ADR antes de implementar.
- Si algo no está documentado: detener y preguntar.
- Si algo contradice la documentación: detener y solicitar ADR.
- Si hay riesgo legal, regulatorio o técnico: detener y advertir.
- ADRs en estado Propuesto: ADR-0007, ADR-0008, ADR-0011, ADR-0017, ADR-0018, ADR-0021.

---

## Rutas relevantes del proyecto

- Backend: `/Volumes/D/vuhmikGO`
- Frontend: `/Volumes/D/vuhmikGO/frontend`
- NVM: `/Volumes/D/nvm`
- Repo: `github.com/Nidael1/VuhmikGO`
- Producción: `app.vuhmik.com` (Hetzner CX23, Falkenstein)

---

## Checklist antes de escribir cualquier código

- [ ] Identifiqué la capa correcta: Core / Shader / Asteroide.
- [ ] La tarea está documentada en los ADRs o sprints.
- [ ] No estoy mezclando responsabilidades entre capas.
- [ ] No voy a editar evidencia emitida.
- [ ] No voy a persistir exports legales.
- [ ] No voy a loguear PHI/PII.
- [ ] Todo error tendrá `error_code`.
- [ ] Toda acción sensible tendrá `reason_code`.
- [ ] La migración es forward-only (sin `.down.sql`).
- [ ] El cambio requiere ADR → lo solicito antes de continuar.

Si algún punto no se cumple: detener ejecución.

---

## Regla final

> La documentación manda. La IA ejecuta. No decide.
> Si no puedo justificar una acción con la documentación existente, no la ejecuto.
