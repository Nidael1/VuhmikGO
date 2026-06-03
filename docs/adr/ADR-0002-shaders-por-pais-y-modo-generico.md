# ADR-0002 — Sistema de Shaders por país y modo genérico

## Estado
Aceptado

## Fecha
2026-06-02

---

## Contexto

VUHMÍK es un solo producto HealthTech oficial, distribuible como SaaS bajo
marca VUHMÍK y como white label sobre el mismo binario Go
(`github.com/Nidael1/VuhmikGO`).

El producto debe poder operar en dos modos sobre el mismo Core agnóstico:

- Como **CRM clínico básico** sin requisitos normativos específicos de país.
- Como **CRM con cumplimiento normativo** de un país determinado.

El cumplimiento normativo (México como primer país) NO puede vivir en el Core.
El Core es agnóstico, determinista e inmutable y no conoce país, profesión,
UI ni reglas legales o clínicas específicas (ver `02_ENGINE_CORE`).

El canon de Shaders (`03_VUHMIK_SHADERS_reglas.md`) ya define:

- `country` como campo opcional del contexto de evaluación (S3, §4.1).
- `extra_shader_keys` como array `0..N` de políticas adicionales (S1/S2).
- Modelo fail-closed: si falta información crítica, la decisión es `deny`.
- Un único `clinical_shader` obligatorio por tenant; `0..1` export shader;
  `0..N` extra shaders.

Este ADR formaliza cómo se expresa la variación por país y el modo genérico
dentro de ese modelo, sin modificar el Core.

---

## Problema

Sin una decisión explícita, existe riesgo de:

- Codificar reglas de cumplimiento de México dentro del Core, violando su
  agnosticismo y obligando a refactor al agregar nuevos países.
- Ambigüedad sobre qué shader representa el "modo básico" versus el "modo
  con cumplimiento", dado que el catálogo actual ya define `med_basic` como
  shader clínico de `medicine`.
- Necesidad de un campo de identidad de país en el tenant que permita
  seleccionar el stack de shaders de forma determinista.

---

## Decisión

### 1. El cumplimiento normativo vive fuera del Core, en la capa Shaders

El Core permanece agnóstico. Toda regla de país o cumplimiento legal se
expresa como Shader consultado por el Core, nunca como lógica del Core.

### 2. El Core trata los keys de Shader como configuración opaca

**El Core NO interpreta los valores `generic_crm`, `mx_medical` ni
`mx_telemedicine_2026`.**

Para el Core, la configuración de shaders de un tenant es **opaca**:

- El Core consulta la frontera Shader de forma determinista y fail-closed,
  pasando el contexto definido en `03_VUHMIK_SHADERS_reglas.md` (S3, §4).
- La selección, compatibilidad, activación, desactivación y administración
  del stack de Shaders pertenece a la capa **Shaders/Admin**, no al Core.
- El Core no tiene tablas ni constantes que enumeren shaders por nombre.
- Cambiar el catálogo de shaders (agregar países, renombrar políticas) no
  requiere cambio en el Core.

### 3. Modo genérico como base: `generic_crm`

Se define `generic_crm` como **modo base clínico mínimo**, sin requisitos
normativos de país.

- Es la configuración por defecto cuando no hay shader de cumplimiento activo.
- Permite operar el producto como CRM clínico simple en cualquier mercado.

### 4. Cumplimiento México inicial: `mx_medical`

Se define `mx_medical` como Shader de cumplimiento normativo para México.

- Cuando está activo, el sistema aplica las reglas de cumplimiento médico
  de México (esquema, workflow, locking, auditoría conforme normativa
  vigente aplicable).
- Cuando NO está activo, el tenant opera en modo `generic_crm` (CRM básico).

### 5. Telemedicina como Shader futuro: `mx_telemedicine_2026`

Se reserva `mx_telemedicine_2026` como Shader **futuro y NO activo**.

- Referencia normativa a validar jurídicamente: "Reforma a la Ley General de
  Salud / salud digital / telemedicina publicada el 15 de enero de 2026".
- El nombre jurídico exacto (ley vs. NOM vs. reglamento) debe confirmarse en
  un ADR posterior antes de activar este shader.
- No se implementa en el alcance actual. El Core queda preparado para
  soportarlo sin refactor por la vía de un nuevo Shader + nuevo Asteroide.

### 6. `country_code` como identidad write-once del tenant

Se reconoce `country_code` como atributo de identidad del tenant.

- Es **write-once**: se fija al crear el tenant y no se modifica después,
  igual que `tenant_area` (S1, §1).
- Determina qué shaders de país son aplicables/compatibles con el tenant.
- Cambiar de país requiere un nuevo tenant, no una mutación.

### 7. `active_shaders` como configuración controlada del tenant

Se formaliza `active_shaders` como la configuración de shaders activos del
tenant, gestionada únicamente desde el Panel Admin (ver `04_ASTEROIDS` y las
reglas del panel: `allow_user_self_manage_shaders` recomendado `false`).

- Respeta el Shader Stack del canon: 1 clinical obligatorio, 0..1 export,
  0..N extra.
- El cumplimiento de país (`mx_medical`) se expresa como parte del stack
  activo del tenant, no como lógica del Core.
- Desde la perspectiva del Core, `active_shaders` es opaco (decisión §2).

### 8. Nuevos países se agregan por Shaders, sin tocar el Core

Agregar un nuevo país (ej. `co_medical`, `es_medical`) consiste en:

- Crear un nuevo Shader de cumplimiento.
- Registrarlo en el catálogo de Shaders.
- Permitir su selección por tenants con el `country_code` correspondiente.

El Core no se modifica. No hay refactor estructural por agregar países.

---

## Alcance respecto al plan de ejecución

Este ADR **no autoriza** mezclar `country_code` ni `active_shaders` con el
Issue #1 (entidades Core base de evidencia) ni con ningún otro issue cuyo
alcance no lo incluya explícitamente.

La disciplina **un issue = un contrato** se mantiene intacta:

- Los atributos `country_code` y `active_shaders` se implementarán **únicamente
  en el issue del plan de ejecución que defina el modelo de tenant /
  configuración**, conforme `12_execution.md`.
- Si dicho issue no existe o no contempla estos campos, se requerirá un issue
  nuevo dedicado, no la ampliación de issues existentes.
- Este ADR fija el **diseño**, no la **secuencia de ejecución**.

---

## Relación con el catálogo existente (tensión de naming a resolver)

El canon actual (`03_VUHMIK_SHADERS_reglas.md`) ya define `med_basic` como
`clinical_shader_key` para `tenant_area: medicine`.

Este ADR introduce `generic_crm` (base) y `mx_medical` (cumplimiento MX).
La relación entre `med_basic` y estos nuevos keys queda fija así:

- **Decisión adoptada:** `mx_medical` se trata como **extra shader de
  cumplimiento** (rol policy/extra), que se suma al `clinical_shader`
  base. El modo `generic_crm` equivale a operar SIN un extra shader de
  cumplimiento de país activo.
- Esto preserva el modelo del canon (1 clinical obligatorio + 0..N extra) y
  cumple el requisito del producto: si se desactiva el shader de cumplimiento,
  el sistema cae a CRM básico de forma fail-safe.
- El mapeo exacto entre `med_basic` y `generic_crm` (si son el mismo clinical
  shader bajo dos nombres comerciales, o si `generic_crm` lo reemplaza como
  key canónico) se resolverá al ejecutar el issue de catálogo de Shaders del
  Sprint 2. No afecta al Core, que trata estos keys como opacos (§2).

---

## Alternativas consideradas

### Codificar reglas de México en el Core
Rechazado. Viola el agnosticismo del Core y obliga a refactor por país.

### Un Core distinto por país (binarios separados por país)
Rechazado. Rompe el modelo de binario único white label y multiplica el
costo operativo para un solo desarrollador.

### Cumplimiento como Asteroide (UX)
Rechazado. El cumplimiento es política, no presentación. Los Asteroides no
deciden reglas clínicas ni legales.

### Que el Core enumere los shaders válidos
Rechazado. Acopla al Core con conocimiento de país y obliga a recompilar al
agregar nuevos shaders. Por eso §2 declara los keys como opacos para el Core.

---

## Consecuencias

- El Core sigue agnóstico y no enumera shaders por nombre.
- El catálogo de Shaders (Sprint 2) registrará `generic_crm`, `mx_medical` y
  reservará `mx_telemedicine_2026` como futuro no activo.
- El Panel Admin (Asteroide de administración) gestiona `active_shaders` con
  `allow_user_self_manage_shaders = false` por defecto.
- El modelo de datos del tenant **debe** contemplar `country_code` (write-once)
  y `active_shaders` (configuración controlada), pero **únicamente** en el
  issue del plan de ejecución que defina ese modelo, sin contaminar otros
  issues.
- Este ADR NO autoriza implementar telemedicina ni activar
  `mx_telemedicine_2026`.
- Este ADR NO autoriza nuevos features fuera del modelo de Shaders descrito.
- Este ADR NO autoriza ampliar el alcance de issues ya definidos.

---

## Documentos impactados

- `03_VUHMIK_SHADERS_reglas.md` (catálogo de shaders de país, Sprint 2).
- `04_ASTEROIDS_productos.md` (Panel Admin gestiona `active_shaders`).
- `12_execution.md` (issue de tenant/configuración debe contemplar
  `country_code` y `active_shaders` cuando llegue su turno; no antes).

---

## Regla final

Este ADR autoriza únicamente el sistema de Shaders por país y el modo
genérico descritos.

No autoriza cambios al agnosticismo del Core, al lifecycle, al modelo de
evidencia append-only, a la disciplina de ejecución (un issue = un contrato),
ni la implementación de telemedicina.

La activación de telemedicina requiere un ADR posterior con el nombre
jurídico exacto de la norma validado.
