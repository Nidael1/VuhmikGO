# ADR-0029 — Stack de Despliegue: Docker + Coolify

## Estado

Aceptado

## Fecha

2026-08-08

## Contexto

`STACK_REAL.md` §3.9 declaraba Docker + Coolify como stack canónico de despliegue. Durante la fase de desarrollo inicial se ejecutó con systemd + nginx sobre Ubuntu bare, sin Docker, y se generó una carpeta `deploy/` con los artefactos correspondientes (vuhmik.service, nginx.conf, deploy.sh). Esa práctica nunca fue formalizada mediante ADR, creando una contradicción entre la documentación canónica y la práctica real.

Este ADR resuelve la contradicción, formaliza la decisión y define las responsabilidades de cada capa.

La decisión se toma en la fase de validación técnica y comercial (§11.0), con el primer despliegue objetivo en el VPS de Hetzner Falkenstein (vuhmik-validacion-01, 2.28.28.18), antes de la migración a Vultr CDMX al ocurrir el Evento de Corte.

## Decisión

**Docker como runtime de la aplicación. Coolify como panel de despliegue.**

El stack canónico queda definido como:

- **Docker Engine** en el servidor, sin Docker Desktop
- **Coolify** (self-hosted) como capa de orquestación, TLS, secrets, logs y healthchecks
- **PostgreSQL** y **Redis** como servicios gestionados por Coolify, corriendo en contenedores
- **El binario Go de VUHMÍK** empaquetado en imagen Docker, desplegado y gestionado por Coolify

La carpeta `deploy/` (systemd + nginx) queda **derogada**. Se conserva en el repositorio como referencia histórica bajo `deploy/legacy/` pero no se usa en ningún entorno activo.

## Responsabilidades por capa

| Responsabilidad | Solución anterior | Solución actual |
|---|---|---|
| Terminación TLS y renovación | certbot manual | Coolify + Traefik automático |
| Gestión de secrets | `.env` en servidor | Variables de entorno en Coolify UI |
| Healthchecks y reinicio | systemd watchdog | Coolify healthchecks |
| Logs de la aplicación | journalctl / archivos | Coolify log viewer |
| Rollback de versión | manual (subir binario) | Coolify UI, un clic |
| Alertas de caída | ninguna | Coolify notificaciones por email |
| Backup de PostgreSQL | backup worker Go (pg_dump) | **Sigue siendo el backup worker Go** |

## Nota crítica sobre backups

Coolify gestiona el despliegue de la aplicación, **no los datos clínicos**. El backup worker de VUHMÍK (`internal/workers/backup.go`) sigue siendo la fuente de verdad para la recuperación de datos. Los datos de PostgreSQL no son responsabilidad de Coolify y deben respaldarse independientemente de él.

Si Coolify falla o se desinstala, los datos clínicos deben ser recuperables sin depender de Coolify.

## Razones para esta decisión

**Docker:**
- Reproducibilidad exacta entre entornos (Mac de desarrollo → Hetzner → Vultr CDMX)
- Rollback de versión de aplicación sin tocar datos
- Aislamiento de dependencias: versiones de Go, PostgreSQL y Redis fijadas en imagen
- Paridad garantizada entre el entorno de validación y el productivo

**Coolify sobre alternativas:**
- UI web para logs, deployments, rollback y secrets: reduce dependencia de SSH en operación diaria
- TLS automático via Traefik sin configuración manual de certbot
- Alertas de caída por email sin desarrollo adicional
- Self-hosted: datos no salen del servidor, cumple con principio de residencia
- Actividad de desarrollo alta en 2026, comunidad activa
- RAM overhead (~600 MB) aceptable en CX23 de 4 GB para fase de validación

**Kamal descartado:** no tiene UI, objetivo declarado de esta fase incluye operación visual.

**systemd + nginx descartado:** sin rollback automático, sin UI de logs, sin alertas.

## Consecuencias

**Positivas:**
- Rollback en 30 segundos desde UI durante demos
- Logs accesibles desde navegador sin SSH
- Alertas automáticas si el servidor cae
- Procedimiento de despliegue idéntico en Hetzner y en Vultr CDMX futuro
- Aprendizaje operativo de Coolify durante fase sin presión de clientes reales

**Negativas y deuda aceptada:**
- Coolify consume ~600 MB de RAM adicional
- Coolify instala su propio PostgreSQL interno; correrán dos instancias en el servidor
- Requiere aprender el modelo de Coolify (proyectos, servicios, recursos)
- `deploy/` actual requiere migración a `deploy/legacy/` y creación de `Dockerfile`

## Pendientes que este ADR desbloquea

- Creación de `Dockerfile` para el binario Go de VUHMÍK
- Instalación de Docker Engine en vuhmik-validacion-01
- Instalación de Coolify en vuhmik-validacion-01
- `RUNBOOK_WAR_MODE.md`: puede escribirse ahora que el stack está definido
- RLS multi-tenant: se implementa en el VPS con usuario `vuhmik_user`

## Documentos afectados

- `STACK_REAL.md` §3.9: ahora formalizado por este ADR
- `deploy/`: contenido movido a `deploy/legacy/`
- `RUNBOOK_WAR_MODE.md`: crear desde cero con procedimiento Coolify

## Referencias

- Coolify v4.x: https://coolify.io
- Docker Engine en Ubuntu: https://docs.docker.com/engine/install/ubuntu/
- §11.0 y §11.2 de documentación canónica
