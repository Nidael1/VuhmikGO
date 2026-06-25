# ADR-0017 — Registro de capacidades por tenant

## Estado
Propuesto

## Fecha
2026-06-24

## Contexto

El Core de VUHMIK es agnostico de dominio (ADR-0016): el mismo motor puede
ser un ECE medico, un ERP o un CRM segun que Shaders y Asteroides se activan.
Lo que define "que es" cada instancia por cuenta es exactamente que modulos
tiene encendidos. Hoy no existe ningun mecanismo que controle esto: todos los
endpoints estan disponibles para cualquier tenant autenticado.

Esto tiene dos consecuencias que hay que resolver antes de construir los
modulos clinicos:

  1. Seguridad: un modulo en desarrollo o pensado para otro rubro podria
     activarse en una cuenta real de produccion sin ningun control.
  2. Comercial: no hay forma de saber que modulos tiene activos cada medico,
     ni de basar la facturacion en esa informacion.

El registro de capacidades resuelve ambas cosas con una sola fuente de verdad,
en dos niveles de escritura con privilegios distintos.

## Decision

### Dos niveles, dos caminos de escritura

El registro de capacidades se organiza en dos tablas con responsabilidades
estrictamente separadas:

  MODULES (plano de control):
    Catalogo global de Shaders y Asteroides que existen en la plataforma.
    Define: id del modulo, rubro (medico/erp/crm), estado de publicacion
    (en_desarrollo / publicado / deprecado) y descripcion.
    Se escribe SOLO por migracion forward-only, desde una maquina de
    confianza, fuera de la web. La app que corre en produccion solo LEE
    esta tabla. No existe ningun endpoint HTTP que escriba en MODULES.

  TENANT_CAPABILITIES (plano de datos):
    Activacion de modulos por cuenta. Para cada par (tenant_id, module_id)
    indica si el modulo esta activo, el plan y el costo.
    Se escribe por el panel de administracion (ADR-0018), pero SOLO sobre
    modulos cuyo publication_status = 'publicado' en MODULES.
    Default: active = false (fail-closed).

### Fail-closed: todo apagado por defecto

Una cuenta nueva no nace con ningun modulo activo. Cada modulo debe ser
activado explicitamente por el admin para ese tenant. Si no hay registro
en TENANT_CAPABILITIES para un par (tenant_id, module_id), el acceso
se niega — igual que si existiera con active = false.

### El Shader como compuerta

Antes de tocar el Core, el Shader ejecuta dos verificaciones en orden:

  1. ¿El modulo existe en MODULES con publication_status = 'publicado'
     y rubro correcto? Si no → niega (modulo no existe o en desarrollo).
  2. ¿TENANT_CAPABILITIES.active = true para este tenant + modulo?
     Si no → niega (modulo no activado para esta cuenta).

Solo si ambas verificaciones pasan, el Shader procede a validar el
contenido y tocar el Core. Esta compuerta es la que convierte el toggle
del admin en una restriccion real de acceso, no en un elemento cosmético.

### Base de facturacion

El mismo registro TENANT_CAPABILITIES es la base del cobro:
  - Los modulos activos definen el plan de la cuenta.
  - El campo `costo` refleja el precio del modulo para esa cuenta.
  - El panel de metricas (ADR-0019) lee esta tabla para calcular MRR,
    churn y distribucion de uso — nunca PHI.

Un solo registro, dos lecturas: el Shader lo lee para seguridad;
el admin lo lee para facturar. No hay desincronizacion posible.

### Super-admin: publicacion por migracion en v1

En v1 no existe panel de super-admin (ADR-0020, diferido). Publicar
un modulo = insertar una fila en MODULES via migracion forward-only.
El rubro medico arranca con los modulos del roadmap clinico pre-publicados
(note, prescription, allergy, diagnosis, immunization, lab_result) como
seed en la migracion inicial de MODULES.

## Dependencias

  - ADR-0016: el Shader consulta MODULES antes de interpretar el type
              del blob opaco del Core.
  - ADR-0018: el panel de toggles es el unico que escribe en
              TENANT_CAPABILITIES (plano de datos).
  - ADR-0019: el panel de metricas lee TENANT_CAPABILITIES como
              fuente de informacion comercial (solo lectura).
  - ADR-0020: el super-admin (diferido) sera el unico que escribe
              en MODULES con UI; en v1 es solo migracion.

## Estado de implementacion

  No implementado.
  Requiere issues de implementacion con:
    - Migracion: tabla modules (id, rubro, publication_status, descripcion,
      created_at). Seed con modulos del rubro medico publicados.
    - Migracion: tabla tenant_capabilities (tenant_id PK, module_id PK FK,
      active bool default false, plan text, costo numeric, updated_at).
    - Puerto CapabilityRepository con metodos:
        IsPublished(moduleID, rubro) bool
        IsActive(tenantID, moduleID) bool
        ListByTenant(tenantID) []Capability
    - Adaptador PostgreSQL del puerto.
    - Integracion en el Shader: verificar IsPublished + IsActive antes
      de procesar cualquier request de modulo clinico.
    - Tests: verificar que fail-closed funciona (sin registro = denegado).

## Consecuencias

  El acceso a cada modulo queda controlado por datos, no por codigo:
  prender o apagar un modulo es un cambio de registro, no un deploy.
  Los modulos en desarrollo son invisibles para el admin y para los
  tenants — no pueden activarse accidentalmente.
  La facturacion y el control de acceso comparten la misma fuente de
  verdad: no hay desincronizacion entre lo que el medico usa y lo que
  se le cobra.
  El Shader gana una verificacion extra antes de tocar el Core; el
  costo es una consulta a PostgreSQL por request de modulo clinico.
  En v1 publicar un modulo requiere correr una migracion — adecuado
  para un equipo de uno; escala con el super-admin UI en v2 (ADR-0020).
