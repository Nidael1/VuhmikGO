# CLEANUP AUDIT — Issue #47

## Fecha
2026-06-04

## Dependencias (go.mod)

Resultado: go mod tidy ejecutado, sin cambios en go.mod.
Todas las dependencias son indirectas de pgx/v5, justificadas:

  github.com/jackc/pgx/v5            — directa, driver PostgreSQL
  github.com/jackc/pgpassfile        — indirecta (pgx)
  github.com/jackc/pgservicefile     — indirecta (pgx)
  github.com/jackc/puddle/v2         — indirecta (pgx pool)
  golang.org/x/sync                  — indirecta (pgx pool)
  golang.org/x/text                  — indirecta (pgx)

Sin dependencias nuevas introducidas. Sin dependencias huérfanas.

## Errores tipados del Core (errors.go / catalog.go)

Resultado: ErrImmutable, ErrInvalidTransition, ErrMissingReasonCode,
ErrInvalidReplacement — todos en uso activo. Sin código muerto.

## Hallazgo: GuardContentEdit sin uso (internal/core/evidence/guard.go)

GuardContentEdit (Issue #8) no tiene llamadas en ningún paquete del
repositorio fuera de su propia definición.

Decisión de este issue: NO eliminar.

Motivo: GuardContentEdit es parte del Core, está documentada como
guard activo en SHADER_FREEZE.md y MVP_HARDENING.md (Issue #46), y
el alcance de Issue #47 prohíbe explícitamente cambios funcionales
al Core. Eliminarla constituye un cambio de contrato del Core, fuera
del alcance de "limpieza de código y dependencias".

Acción recomendada: si se confirma que GuardContentEdit es
redundante con GuardMutation, su eliminación requiere un ADR y un
issue propio que actualice el Core, SHADER_FREEZE.md y
MVP_HARDENING.md de forma consistente.

## Resultado

Build limpio, vet limpio, gofmt limpio, tests verdes (sin cambios).
Sin código muerto eliminado en este issue (el único candidato
identificado pertenece al Core bajo freeze).
