# VUHMÍK — Registro de Issues

Inventario de issues del proyecto: lo construido y lo pendiente.

Nota de procedencia (para que sepas qué tan fino es cada dato):
- Los issues #111 a #123 están detallados desde el historial de commits de esta
  sesión (título y alcance reales).
- Los issues #1 a #110 (MVP y Sprint 7) se describen a nivel de bloque y
  entregable; los títulos granulares de cada uno viven en el tracker del repo.
- Los pendientes se derivan de los ADR y van marcados como propuestos.

Última actualización: 2026-08-19

---

## Estado general

  Completados:   MVP (#1-#66) + Sprint 7 (#67-#112) + Sprint 8 (#113-#123) + Sprint 9.1-9.4 (#124-#155) + #156-#162
  ADR:           ADR-0001 a ADR-0030 redactados y aceptados
  Pendientes:    ninguno en sprint 9.5

  Tests:         24 pasando · go build ok · go vet ok
  Rama:          main

---

# Completados

## MVP — Núcleo Engine + Shaders + Evidencia (#1-#66)

Bloque fundacional. Entregables (títulos granulares en el repo):

  - Core de evidencia inmutable: entidad con subject + contenido, estados.
  - Lifecycle draft → issued → locked/voided con guardas de mutación.
  - Void + replace (corrección sin borrado), cadena replaced_by.
  - Shaders: MedicalBasicShader y LegalExportShader (única vía al Core).
  - Puerto EvidenceRepository + adaptadores inmemory y postgres.
  - Export legal efímero (sin persistencia, Cache-Control: no-store).
  - Observabilidad: logger slog JSON, métricas agregadas, sin PHI/PII.
  - Multi-tenant fail-closed (aislamiento por tenant verificado).
  - Migraciones forward-only (sin .down.sql).

## Sprint 7 — API JSON + Frontend Vue + Pacientes (#67-#112)

Entregables de bloque:

  - API JSON /api/v1 (router dinámico, DI, helpers writeJSON/writeError).
  - Autenticación JWT + middleware de contexto tenant/actor.
  - Frontend Vue 3 + Vite + TS + Pinia + Vue Router (SPA).
  - Repositorios e infraestructura HTTP del frontend (auto-refresh).
  - Módulo de pacientes NOM-004 (tabla patients, expediente EXP-NNNN).
  - Vistas: login/registro, lista de pacientes, detalle (hoja continua),
    nueva nota vinculada al paciente, edición fluida (ADR-0006).

  Detallados:
  #111  patient list search counter uppercase name
        (búsqueda, contador, nombres en MAYÚSCULAS)
  #112  patient list sort A-Z / expediente
        (botones de orden alfabético y por número de expediente)

## Sprint 8.2 — Seguridad de sesión (#113-#116)

  #113  refresh tokens migration stateful sessions
        (tabla refresh_tokens en PostgreSQL, solo hash)
  #114  jwt 15min + refresh tokens stateful postgres
        (access token 15 min; refresh 7 días, stateful)
  #115  auth refresh endpoint token rotation frontend
        (POST /api/v1/auth/refresh con rotación; auto-refresh en 401)
  #116  auth logout refresh token revocation
        (POST /api/v1/auth/logout; revocación real de sesión)

## Sprint 9.5 — Corrección de bugs de integridad (#156-#161) — 2026-08-19

  #156  bugfix patch parcial edicion inline paciente race condition
        (saveName/saveSexo solo envían su campo; PatientRequest opcionales;
         selector reabre al fallar API)
  #157  bugfix doble request @change+@blur en selector de sexo
        (elimina @blur; solo @change dispara el guardado)
  #158  bugfix guard null en saveSexo antes de non-null assertion
        (captura patient.value en variable local antes del try)
  #159  bugfix mensaje de error incorrecto para fecha con formato invalido
        (separa err!=nil de fn.After; mensaje YYYY-MM-DD vs "no puede ser futura")
  #160  bugfix HandlePatientUpdate rechaza sexo invalido, antes devolvía 200 OK
        (si sexo no es M/F/I y viene en el payload, responde 400 INVALID_FIELDS)
  #161  bugfix fecha maxima del picker en hora local no UTC
        (localDateToday() con getFullYear/getMonth/getDate en PatientNewView)
  #162  validacion demografica de paciente al Shader med_basic (ADR-0030)
        (ValidatePatientDemographics en medical_basic.go; ER-SHADER-010;
         handlers delegan al Shader; ADR-0030 aceptado)

---

## Sprint 8.3 — WAR-A + integridad + preparación IPS (#117-#123)

  #117  redis integration war-a broker
        (cliente go-redis/v9; REDIS_URL en secretos validados)
  #118  backup worker postgresql 24h purge 7days
        (pg_dump cada 24h; purga de backups +7 días; shutdown graceful)
  #119  metrics purge worker 30days retention war-a
        (reset de contadores agregados cada 30 días)
  #120  golive checklist estado actual verificado
        (docs/GOLIVE_CHECKLIST.md con estado vs WAR-A)
  #121  sha256 hash evidencia export adr-0008
        (paquete integrity; hash canónico en el export)
  #122  export xml adr-0007 json xml format selector
        (XML HL7 CDA simplificado; selección por header Accept)
  #123  curp users patients migration handlers frontend
        (CURP en users —único global— y patients —opcional por tenant)

---

# Decisiones de arquitectura (ADR)

  Redactados:   ADR-0001 a ADR-0015 (en main)
  Pendientes:   ADR-0016 a ADR-0020

  Detalle en REGISTRO_ADR.md. Varios ADR generan issues de implementación
  (listados abajo como pendientes).

---

# Pendientes de implementación

## Fase 8.1 — Infraestructura (al hacer cutover al VPS)

  #124  Dockerfile + docker-compose
  #125  Coolify setup en VPS (Hetzner)
  #126  TLS + dominio + firewall (22/80/443) + SSH hardening

  Bloqueados hasta el cutover; el desarrollo no los necesita.

## Sprint 9 — Core agnóstico (ADR-0016) [propuesto]

  - Migración forward-only: content opaco (envolver notes en blob).
  - Sacar la semántica médica (Notes) del Core; SubjectID → clave opaca.
  - type del registro vive DENTRO del blob; lo lee el Shader.
  - Hash (ADR-0008) recalculado sobre el blob opaco + metadata.
  - Reinterpretar ADR-0011 a 0015 sobre el modelo de contenido opaco.

## Capacidades + administración (ADR-0017/0018/0019) [propuesto]

  - Registro de capacidades (catálogo MODULES + activación por tenant),
    fail-closed; el Shader lo consulta antes de tocar el Core.
  - Flag de publicación + rubro (escrito solo por migración; off-web).
  - Panel admin de toggles (escribe activación por cuenta; solo publicado).
  - Bandera is_admin; suspensión que bloquea login sin tocar datos.
  - Panel de métricas (solo lectura): lista de doctores, conteos, MRR,
    churn, uso de módulos; agregados, nunca PHI; precálculo por worker.

## Módulos clínicos IPS (ADR-0011 a 0015) [propuesto]

  Cada módulo = un mini-sprint con: migración (contenido tipado sobre el
  registro genérico), validación en Shader, handlers API draft/emit/void,
  endpoint de listado por paciente, proyección a sección IPS, y frontend
  en el detalle del paciente.

  - Medicación / receta electrónica (ADR-0011) — sección IPS obligatoria.
  - Alergias e intolerancias (ADR-0012) — sección IPS obligatoria.
  - Diagnósticos / lista de problemas CIE-10 (ADR-0013) — obligatoria.
  - Inmunizaciones / vacunación (ADR-0014) — recomendada.
  - Resultados de laboratorio (ADR-0015) — recomendada.

## Super-admin (ADR-0020) [diferido]

  - Control de catálogo, publicación y rubro como app interna aislada
    (localhost/VPN), nunca en la web pública. No se construye en v1.

---

# Orden sugerido de ejecución

  1. ADR-0016 a 0019 (redactar) → cierra la base del Core agnóstico y admin.
  2. Sprint 9: Core agnóstico (ADR-0016) sobre el modelo de contenido opaco.
  3. Capacidades + paneles admin (ADR-0017/0018/0019).
  4. Módulos clínicos IPS, uno por mini-sprint (receta primero).
  5. Fase 8.1 (infra) al cutover.
  6. ADR-0020 (super-admin) cuando escale.

Disciplina: un issue = una rama = un commit = un PR = un merge.
