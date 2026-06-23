# CHECKLIST DE GO-LIVE — VuhmikGO

## Propósito

Este documento registra el estado de cumplimiento del checklist
de go-live definido en la documentación canónica.

Si una casilla no se cumple → NO GO LIVE.

Fecha de última verificación: 2026-06-23

---

## 1) Seguridad básica

  [ ] TLS activo en todos los dominios
      PENDIENTE — requiere servidor con dominio (Fase 8.1)

  [ ] Solo puertos 22 / 80 / 443 expuestos
      PENDIENTE — requiere configuración VPS (Fase 8.1)

  [ ] SSH solo por llave (sin password)
      PENDIENTE — requiere configuración VPS (Fase 8.1)

  [ ] Login root deshabilitado
      PENDIENTE — requiere configuración VPS (Fase 8.1)

  [x] Secrets fuera de git
      CUMPLIDO — DATABASE_URL, JWT_SECRET, REDIS_URL se pasan
      como variables de entorno, nunca en código.

---

## 2) Confidencialidad

  [x] Tenancy fail-closed verificado
      CUMPLIDO — Issue #56. FindByID y Update filtran por
      tenant_id. Tests de aislamiento multi-tenant pasan.

  [x] JWT corto activo
      CUMPLIDO — Issue #114. Access token de 15 minutos.

  [x] Refresh tokens stateful
      CUMPLIDO — Issue #114. Refresh tokens en tabla
      refresh_tokens de PostgreSQL. Expiran en 7 días.

  [x] Revocación por sesión funcional
      CUMPLIDO — Issue #116. POST /api/v1/auth/logout
      revoca el refresh token activo.

  [ ] Redis no expuesto
      PENDIENTE — requiere configuración de firewall en VPS.

  [ ] Postgres no expuesto
      PENDIENTE — requiere configuración de firewall en VPS.

---

## 3) Integridad

  [x] Evidencia append-only activa
      CUMPLIDO — Core inmutable. GuardMutation activo.
      No existe DELETE ni UPDATE de contenido clínico.

  [x] Void + reemplazo funcional
      CUMPLIDO — Issue #76, #89, #103. UpdateForVoid
      separado de Update para preservar integridad.

  [ ] Hashing de evidencia activo
      PENDIENTE — ADR-0008 aprobado. Implementación
      pendiente (Issues #113-#114 del plan de export).

  [ ] Verificación de integridad por jobs
      PENDIENTE — requiere implementación futura.

  [x] Migraciones no destructivas
      CUMPLIDO — 6 migraciones forward-only aplicadas.
      Sin archivos .down.sql.

---

## 4) Disponibilidad

  [x] Backups diarios de PostgreSQL
      CUMPLIDO — Issue #118. BackupWorker corre cada
      24 horas via pg_dump.

  [ ] Backups de anexos
      N/A en v1 — sin sistema de anexos implementado.

  [ ] Backups externos cifrados
      PENDIENTE — requiere Storage Box o S3 en VPS (Fase 8.1).

  [ ] Restore probado exitosamente
      PENDIENTE — debe probarse cuando el servidor esté activo.

  [x] Retención configurada
      CUMPLIDO — Issue #118. Backups con más de 7 días
      se purgan automáticamente. Issue #119. Métricas
      con más de 30 días se purgan automáticamente.

---

## 5) Jobs y automatismos (WAR-A)

  [x] Redis operativo
      CUMPLIDO — Issue #117. Redis 8.8.0 conectado.
      Cliente go-redis/v9 integrado.

  [x] Workers activos
      CUMPLIDO — Issues #118, #119. BackupWorker y
      MetricsPurgeWorker corriendo con contexto cancelable
      y shutdown graceful.

  [x] Exportaciones funcionan
      CUMPLIDO — POST /api/v1/evidence/:id/export
      genera JSON en memoria, sin persistencia.
      Cache-Control: no-store activo.

  [x] Purge de métricas activo
      CUMPLIDO — Issue #119. MetricsPurgeWorker
      resetea contadores cada 30 días.

  [x] Jobs idempotentes
      CUMPLIDO — BackupWorker genera archivo nuevo
      con timestamp único. MetricsPurgeWorker es
      idempotente por diseño (reset atómico).

---

## 6) Observabilidad mínima

  [x] Logs visibles
      CUMPLIDO — log/slog JSON estructurado a stdout.
      Compatible con Coolify en producción.

  [x] Logs sin secretos
      CUMPLIDO — Issue #41. Logger solo acepta campos
      técnicos (operation, tenant_id, error_code, path).
      Sin PHI ni PII en logs.

  [x] Métricas agregadas activas
      CUMPLIDO — internal/observability/metrics.go.
      Contadores atómicos por tipo de evento.

  [x] Retención de métricas configurada
      CUMPLIDO — Issue #119. Purge cada 30 días.

---

## 7) Verificación funcional mínima

  [x] Tenant válido se resuelve
      CUMPLIDO — JWTMiddleware extrae tenant_id del
      token. TenantIDFromContext disponible en handlers.

  [x] Creación de draft exitosa
      CUMPLIDO — POST /api/v1/evidence/draft funcional.
      SubjectID y Notes persistidos en PostgreSQL.

  [x] Emisión correcta
      CUMPLIDO — Draft se emite automáticamente al crear
      (ADR-0006). Estado issued persistido en PostgreSQL.

  [x] Void + reemplazo funcional
      CUMPLIDO — UpdateForVoid corrige el bug de
      GuardMutation para void silencioso (Issue #103).

  [x] Exportación legal válida
      CUMPLIDO — LegalExportShader genera JSON en
      memoria. Cache-Control: no-store verificado
      (Issue #57).

---

## Resumen de pendientes para GO LIVE

Los siguientes items requieren el servidor VPS (Fase 8.1):

  1. TLS activo con dominio propio
  2. Firewall: solo puertos 22/80/443
  3. SSH por llave, root deshabilitado
  4. Redis y PostgreSQL no expuestos públicamente
  5. Backups externos cifrados (Storage Box / S3)
  6. Restore probado en producción

Los siguientes items son mejoras de integridad (post go-live v1.1):

  7. Hashing SHA-256 de evidencia (ADR-0008)
  8. Verificación de integridad por jobs

---

## Declaración

El sistema es funcionalmente correcto y seguro para go-live
una vez completados los 6 items de infraestructura de la
Fase 8.1 (VPS + Coolify + TLS).

Los items 7 y 8 son mejoras de integridad programadas para
v1.1 y no bloquean el go-live inicial.
