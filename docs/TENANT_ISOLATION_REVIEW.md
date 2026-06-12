# TENANT ISOLATION REVIEW — Issue #56

## Fecha
2026-06-04

## Resultado
Vulnerabilidad encontrada y corregida.

## Hallazgo

EvidenceRepository.FindByID(id) y Update(e) operaban solo por ID,
sin filtrar por tenant_id. Un tenant que conociera o adivinara el
ID de un registro de otro tenant podia leerlo o mutarlo.

## Correccion aplicada

  - ports.EvidenceRepository: FindByID y Update ahora requieren
    tenantID explicito.
  - internal/infrastructure/inmemory: filtra por tenant_id,
    retorna "no encontrado" si el tenant no coincide (no distingue
    entre "no existe" y "es de otro tenant").
  - internal/infrastructure/postgres: agrega "AND tenant_id = $N"
    en WHERE de SELECT y UPDATE.
  - internal/application/ece_service.go: Issue, Lock, IssueAndLock,
    Void y Replace ahora exigen tenantID y lo propagan al repo.

## Tests de aislamiento agregados

  internal/application/ece_service_test.go
    - TestAislamientoMultiTenant_FindByID: tenant-B no puede leer
      un registro de tenant-A. PASS.
    - TestAislamientoMultiTenant_Issue: tenant-B no puede emitir
      un registro de tenant-A; tenant-A si puede. PASS.

## Estado de tests del proyecto

  Core: 9 PASS, Shaders: 13 PASS, Application: 2 PASS = 24 total.

## Hallazgo adicional (fuera de alcance de #56, NO corregido aqui)

ECEService.Void esta roto para ambos estados validos de origen:

  - Sobre un registro "issued": repo.Update llama GuardMutation,
    que bloquea cualquier update si el estado actual es issued o
    locked — incluyendo la transicion legitima a voided.
  - Sobre un registro "draft": evidence.Void llama GuardTransition,
    y draft->voided no esta en la matriz de lifecycle (solo
    issued->voided y locked->voided son validos).

Este es un bug de lifecycle preexistente (Issues #9/#36/#39), no
relacionado con aislamiento multi-tenant. Requiere su propio issue:
el repositorio debe permitir la mutacion issued/locked->voided
especificamente para la operacion Void (y Replace), sin relajar
GuardMutation para edicion de contenido arbitraria.

## Cambios de contrato

ports.EvidenceRepository.FindByID y Update cambian de firma
(agregan tenantID). Esto es un ajuste de la frontera de
infraestructura para cumplir la regla ya vigente de aislamiento
multi-tenant fail-closed; no introduce nueva logica de negocio ni
modifica contratos Core/Shader.
