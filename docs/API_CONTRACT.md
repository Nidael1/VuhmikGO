# CONTRATO REST JSON — /api/v1

## Version
1.0 — 2026-06-12

## Base URL

  /api/v1

## Autenticacion

  Todas las rutas protegidas requieren:
    Authorization: Bearer <jwt_token>

  Rutas publicas (sin token):
    POST /api/v1/auth/register
    POST /api/v1/auth/login

## Formato de respuesta

  Toda respuesta sigue el mismo envelope:

  Exito:
    {
      "data": { ... },
      "error": null
    }

  Error:
    {
      "data": null,
      "error": {
        "code": "ERROR_CODE",
        "message": "descripcion legible"
      }
    }

## Formato de fechas

  ISO 8601 UTC: 2026-06-12T10:00:00Z

## Codigos de error de aplicacion

  INVALID_JSON         — payload no parseable
  MISSING_FIELDS       — campos obligatorios ausentes
  INVALID_CREDENTIALS  — email o password incorrectos
  EMAIL_EXISTS         — email ya registrado
  UNAUTHORIZED         — token ausente o invalido
  INVALID_TOKEN        — token expirado o malformado
  NOT_FOUND            — recurso no encontrado
  FORBIDDEN            — operacion no permitida para este tenant
  INTERNAL_ERROR       — error interno del servidor

  Errores del Core (mapeados desde error_code):
  EVIDENCE_IMMUTABLE        — ER-CORE-001
  EVIDENCE_INVALID_TRANSITION — ER-CORE-002
  EVIDENCE_MISSING_REASON   — ER-CORE-003
  EVIDENCE_INVALID_REPLACE  — ER-CORE-004

## Endpoints de autenticacion

  POST /api/v1/auth/register
    Request:  { "email": string, "password": string }
    Response: { "token": string, "tenant_id": string, "actor_id": string }

  POST /api/v1/auth/login
    Request:  { "email": string, "password": string }
    Response: { "token": string, "tenant_id": string, "actor_id": string }

  GET /api/v1/auth/me
    Headers:  Authorization: Bearer <token>
    Response: { "actor_id": string, "tenant_id": string }

## Endpoints de evidencia (Sprint 7.1)

  GET    /api/v1/evidence
    Response: { "items": [ EvidenceItem ] }

  GET    /api/v1/evidence/:id
    Response: EvidenceItem

  POST   /api/v1/evidence/draft
    Request:  { "subject_id": string, "notes": string }
    Response: EvidenceItem

  POST   /api/v1/evidence/:id/emit
    Response: EvidenceItem

  POST   /api/v1/evidence/:id/void
    Request:  { "reason_code": string }
    Response: EvidenceItem

  POST   /api/v1/evidence/:id/replace
    Request:  { "reason_code": string, "replacement_id": string }
    Response: { "voided": EvidenceItem, "replacement": EvidenceItem }

  POST   /api/v1/evidence/:id/export
    Response: JSON attachment (Content-Disposition: attachment)

## Tipo EvidenceItem

  {
    "id":              string,
    "tenant_id":       string,
    "state":           "draft" | "issued" | "locked" | "voided",
    "created_at":      string (ISO 8601),
    "issued_at":       string | null,
    "voided_at":       string | null,
    "replaced_by_id":  string | null
  }

## Status codes HTTP

  200 OK              — lectura exitosa
  201 Created         — recurso creado
  400 Bad Request     — payload invalido o campos faltantes
  401 Unauthorized    — token ausente o invalido
  403 Forbidden       — operacion no permitida
  404 Not Found       — recurso no encontrado
  409 Conflict        — recurso ya existe
  422 Unprocessable   — validacion de negocio fallida
  500 Internal Error  — error interno del servidor

## Reglas del contrato

  1. El tenant_id nunca viene del cliente — siempre del JWT.
  2. El actor_id nunca viene del cliente — siempre del JWT.
  3. Todo error tiene un code string, nunca solo mensaje libre.
  4. Las fechas son siempre UTC.
  5. El campo data es null en errores, el campo error es null en exitos.
  6. No existe paginacion en v1 — se agrega en v2 si aplica.
