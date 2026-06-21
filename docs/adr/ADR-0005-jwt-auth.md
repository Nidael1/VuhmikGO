# ADR-0005 — Autenticacion JWT propia para API post-MVP

## Estado
Aceptado

## Fecha
2026-06-12

## Contexto

El MVP v0.1.0-rc1 usa headers X-Tenant-ID y X-Actor-ID de confianza
del cliente. Esto es inseguro para produccion: cualquier cliente puede
suplantar un tenant o actor. El ciclo post-MVP requiere autenticacion
real para que la SPA Vue pueda operar de forma segura.

## Decision

Se implementa autenticacion JWT propia en Go.

El backend genera y valida JWT firmados con HMAC-SHA256.
El token contiene tenant_id y actor_id resueltos por el servidor.
El cliente Vue recibe el token en login y lo envia en cada request
como Bearer token en el header Authorization.
El middleware JWT reemplaza X-Tenant-ID y X-Actor-ID en rutas /api/v1.

Los headers X-Tenant-ID y X-Actor-ID se conservan para las rutas
HTML historicas del MVP (compatibilidad hacia atras).

## Endpoints de autenticacion

  POST /api/v1/auth/register  — registro de medico (email + password)
  POST /api/v1/auth/login     — login, retorna JWT
  GET  /api/v1/auth/me        — perfil del usuario autenticado

## Estructura del JWT

  Header: HS256
  Payload:
    sub        — user_id
    tenant_id  — tenant del medico
    actor_id   — identificador del actor (igual a sub en MVP)
    exp        — expiracion (24h por defecto)
    iat        — emitido en

## Almacenamiento en frontend

  El token JWT se almacena en memoria (Pinia store).
  No se usa localStorage ni sessionStorage en primera instancia.
  Si el usuario recarga, se redirige a login.
  En iteraciones futuras puede evaluarse httpOnly cookie.

## Tabla requerida

  users (ver migracion 000003_create_users.up.sql)
    id          TEXT NOT NULL (PK)
    tenant_id   TEXT NOT NULL
    email       TEXT NOT NULL UNIQUE
    password_hash TEXT NOT NULL
    created_at  TIMESTAMPTZ NOT NULL

## Alternativas consideradas

  Auth0 / Clerk — rechazado: dependencia externa, costo, complejidad
  para MVP de presentacion a VC.
  httpOnly cookies — evaluable en iteracion futura. JWT en memoria
  es suficiente para demo controlada.

## Consecuencias

  Se agrega tabla users con migracion forward-only.
  Se agrega dependencia golang-jwt/jwt para Go.
  El middleware JWT resuelve tenant_id y actor_id desde el token.
  La SPA Vue no decide quien es el tenant — el backend lo resuelve.
