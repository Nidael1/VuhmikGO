# ADR-0021 — Perfil profesional por rubro

## Estado
Propuesto

## Fecha
2026-06-27

## Contexto

La tabla `users` contiene credenciales y sesion del actor autenticado:
email, password_hash, tenant_id, curp. Es agnóstica de dominio — sirve
para cualquier rubro (medicina, nutricion, legal, notarial).

Para emitir una receta electronica valida bajo NOM-024-SSA3-2012, el
sistema necesita datos profesionales del medico: nombre completo, cedula
profesional y especialidad. Estos datos no pertenecen a `users` porque:

  1. Son especificos de un rubro (medicina en Mexico). Un nutriologo
     tiene cedula pero no especialidad medica. Un notario tiene datos
     distintos. Un ERP no tiene ninguno de estos campos.
  2. Colocarlos en `users` acoplaría el Core a un dominio especifico,
     violando ADR-0016 (Core agnostico).
  3. El mismo actor podria operar en mas de un rubro en el futuro;
     sus credenciales son las mismas pero su perfil profesional cambia.

## Decision

### Tabla `professional_profiles`

Se crea una tabla separada para el perfil profesional del actor,
indexada por usuario y rubro:

  professional_profiles:
    user_id        text NOT NULL FK → users.id
    tenant_id      text NOT NULL
    rubro          text NOT NULL  (medicine / nutrition / legal / generic)
    nombre_completo text NOT NULL DEFAULT ''
    cedula_profesional text NOT NULL DEFAULT ''
    especialidad   text NOT NULL DEFAULT ''
    datos_extra    jsonb          (campos adicionales por rubro, nullable)
    created_at     timestamptz NOT NULL DEFAULT now()
    updated_at     timestamptz NOT NULL DEFAULT now()
    PK: (user_id, rubro)

### Separacion de responsabilidades

  users                  Credenciales y sesion. Agnostica. No cambia.
  professional_profiles  Perfil del actor por rubro. Especifica de dominio.
                         La conoce el Shader, no el Core.

### Rubro inicial: medicine

En v1 solo existe el perfil de rubro `medicine`. Los campos
`cedula_profesional` y `especialidad` son obligatorios para emitir
una receta (NOM-024) pero opcionales para registrarse — el medico
los completa en su perfil antes de emitir la primera receta.

### Acceso y modificacion

El actor solo puede leer y modificar su propio perfil.
El Shader de receta verifica que `cedula_profesional` y `especialidad`
esten presentes antes de permitir la emision (draft → issued).
Si faltan, el Shader niega con error tipado que la UI traduce a:
"Completa tu perfil profesional antes de emitir una receta."

### API

  GET  /api/v1/profile          → devuelve perfil del actor autenticado
  PUT  /api/v1/profile          → actualiza perfil del actor autenticado

El perfil se crea automaticamente (vacio) al registrar un usuario nuevo,
con rubro = 'medicine' en v1.

### Frontend

Pantalla de perfil accesible desde el menu lateral. Campos editables:
nombre completo, cedula profesional, especialidad. Sin modal — edicion
directa en la pantalla. Guardado via PUT /api/v1/profile.

## Dependencias

  - ADR-0016: professional_profiles es capa de dominio, no Core.
              El Core no conoce ni referencia esta tabla.
  - ADR-0011: el Shader de receta verifica cedula + especialidad antes
              de permitir emision. Sin perfil completo no hay receta.
  - ADR-0017: el rubro del perfil debe coincidir con el rubro del modulo
              activado en tenant_capabilities.

## Estado de implementacion

  No implementado.
  Requiere issues de implementacion:
    - Migracion 000010: tabla professional_profiles + creacion automatica
      de perfil vacio al registrar usuario (rubro medicine por defecto).
    - Puerto ProfileRepository: Get(userID) y Update(userID, data).
    - Adaptador PostgreSQL del puerto.
    - Handlers GET /profile y PUT /profile.
    - Auth/me actualizado para incluir datos del perfil.
    - Frontend: pantalla de perfil en menu lateral.
    - Shader de receta (ADR-0011): verificacion de cedula + especialidad
      antes de emision.

## Consecuencias

  El Core permanece agnostico: ninguna tabla del Core gana columnas
  de dominio medico.
  La receta puede cumplir NOM-024 sin contaminar la capa de usuarios.
  El modelo escala a otros rubros sin migraciones destructivas: agregar
  un nutriologo es insertar un perfil con rubro = 'nutrition'.
  El campo datos_extra JSONB absorbe variaciones de rubro sin romper
  el esquema base.
  El medico debe completar su perfil antes de emitir su primera receta;
  esto es un requisito legal, no una restriccion arbitraria.
