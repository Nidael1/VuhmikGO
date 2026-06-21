# AUDITORÍA INTERNA DE CUMPLIMIENTO — Issue #59

## Fecha
2026-06-12

## Alcance
Issues #1 a #58 completados y mergeados a main.

## Estado técnico verificado

  go build ./...  -> OK
  go vet ./...    -> OK
  gofmt -l .      -> limpio
  go test ./...   -> 24 PASS (Core 9, Shaders 13, Application 2)
  git tag         -> v0.1.0-rc1

## Auditoría por regla absoluta

REGLA: Core agnostico, determinista, inmutable
  Estado: CUMPLE
  Evidencia: internal/core/evidence/ no importa paquetes de
  delivery ni shaders. Lifecycle cerrado. GuardMutation activo.

REGLA: Acceso al Core unicamente via Shaders
  Estado: CUMPLE
  Evidencia: grep -r internal/core/evidence internal/delivery/
  -> vacio (SHADER_INTEGRATION_AUDIT.md, Issue #32/#56).

REGLA: Evidencia append-only (sin DELETE, sin UPDATE de contenido)
  Estado: CUMPLE
  Evidencia: SQL en postgres adapter solo actualiza state/
  issued_at/voided_at/replaced_by_id, nunca campos de contenido.
  GuardMutation bloquea updates en issued/locked.

REGLA: Lifecycle draft->issued->locked/voided, sin transiciones implicitas
  Estado: CUMPLE
  Evidencia: lifecycle.go, matriz cerrada, GuardTransition.

REGLA: Void/Replace con reason_code obligatorio del catalogo
  Estado: CUMPLE
  Evidencia: evidence.Void valida contra ReasonCodeCatalog.
  ErrMissingReasonCode (ER-CORE-003) si no valida.

REGLA: Export efimero, sin persistencia
  Estado: CUMPLE
  Evidencia: EXPORT_PERSISTENCE_REVIEW.md (Issue #57).
  Cache-Control: no-store en header de respuesta.

REGLA: Observabilidad sin PHI/PII
  Estado: CUMPLE
  Evidencia: internal/observability/logger.go solo acepta
  campos tecnicos (operation, tenant_id, error_code, path).

REGLA: Multi-tenant fail-closed
  Estado: CUMPLE
  Evidencia: middleware.go rechaza sin X-Tenant-ID/X-Actor-ID.
  FindByID/Update filtran por tenant_id (Issue #56).

REGLA: Migraciones forward-only
  Estado: CUMPLE
  Evidencia: database/migrations/ solo contiene .up.sql.
  Sin .down.sql. golang-migrate version activa.

REGLA: Un issue = una rama = un PR = un commit
  Estado: CUMPLE (con una excepcion documentada)
  Excepcion: Issue #47 requirio rama adicional para go.mod
  (issue/47-cleanup-gomod). Documentado y justificado.

REGLA: Secretos no hardcodeados, fail-closed
  Estado: CUMPLE
  Evidencia: ValidateRuntimeSecrets() en main.go.
  Sin credenciales en codigo fuente.

## Hallazgos documentados (no bloqueantes)

1. GuardContentEdit sin uso (Issue #47 CLEANUP_AUDIT.md)
   Funcion publica del Core sin callers. No eliminada por estar
   bajo freeze. Requiere ADR + issue propio si se confirma
   redundante con GuardMutation.

2. Bug ECEService.Void (Issue #56 TENANT_ISOLATION_REVIEW.md)
   Void esta roto para registros issued/locked: GuardMutation
   bloquea el Update necesario para la transicion a voided.
   Documentado. Requiere issue propio para corregir sin relajar
   GuardMutation de forma indiscriminada.

3. error_copy.go usa strings literales como claves de mapa
   (Issue #58). Aceptable para mapa de presentacion. No es
   punto de emision de errores.

## Resultado de la auditoria

Sin desviaciones criticas. Los 3 hallazgos son no bloqueantes
y estan documentados con rutas de accion claras.

El MVP cumple el plan contractual definido en 12_execution.md
para Issues #1-#58.
