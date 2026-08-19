# ADR-0030 — Validación demográfica de paciente en Shader med_basic

## Estado
Aceptado

## Fecha
2026-08-19

## Contexto

Al implementar la validación "fecha de nacimiento no puede ser futura" (issue #159,
sprint 9.5), esta regla quedó escrita directamente en los handlers HTTP
`HandlePatientCreate` y `HandlePatientUpdate` (`internal/delivery/http/api/patient_handlers.go`).

El CLAUDE.md es explícito:

  "NOM-024-SSA3-2012: estándar de información clínica. Las validaciones deben
   vivir en el Shader `med_basic`, no en handlers."

Una validación demográfica clínica (fecha de nacimiento, sexo biológico) es
exactamente el tipo de regla que pertenece al Shader: es una restricción de
dominio clínico, no una restricción de protocolo HTTP. Si en el futuro otra
ruta de ingreso (batch importer, patient transfer, API interna) llama la
lógica sin pasar por estos handlers, la validación se salta silenciosamente.

Adicionalmente, el Shader `med_basic` actual solo valida el contexto de
operación (tenant_id, actor_id, operation) pero no los datos demográficos del
paciente, dejando ese contrato incompleto respecto a lo que declara NOM-024.

## Decisión

### Dónde vive la validación

Las reglas clínicas sobre datos demográficos del paciente viven en
`internal/shaders/medical_basic.go`, en una función exportada
`ValidatePatientDemographics(nombre, fechaNacimiento, sexo string) error`.

Los handlers HTTP llaman esta función antes de persistir. Si retorna error,
el handler responde 400 con el `error_code` correspondiente. La función no
conoce HTTP ni PostgreSQL — solo evalúa strings y devuelve un error tipado.

### Contrato de la función

```go
// ValidatePatientDemographics valida los datos demográficos básicos de un
// paciente conforme a NOM-024-SSA3-2012.
// Retorna nil si los datos son válidos.
// Retorna un error con código ER-SHADER-010 si alguna regla clínica falla.
func ValidatePatientDemographics(nombre, fechaNacimiento, sexo string) error
```

Reglas implementadas:
  1. Si `fechaNacimiento` no está vacío, debe tener formato `YYYY-MM-DD`.
  2. Si `fechaNacimiento` no está vacío, no puede ser una fecha futura
     (comparado contra fecha UTC del servidor).
  3. Si `sexo` no está vacío, debe ser `M`, `F` o `I`.

Las validaciones son opcionales por campo (aplican solo si el campo viene
en el payload) para soportar actualizaciones parciales (PATCH semántico
sobre el endpoint PUT existente, per ADR-0030 y issue #156).

### Error code

  ER-SHADER-010  validación demográfica de paciente fallida

Catalogado en el catálogo de códigos de error del proyecto.

### Lo que NO cambia

  - El Core no conoce esta validación. Solo el Shader la ejecuta.
  - Los handlers siguen siendo la capa de entrega; solo delegan la decisión.
  - El lifecycle draft → issued → locked/voided no cambia.
  - No se introduce ningún estado nuevo ni migración.

## Dependencias

  - ADR-0002: Shaders por país y modo genérico — esta decisión es consecuencia
              directa de ese principio.
  - ADR-0003: Estructura Go — la función vive en `internal/shaders/`.
  - NOM-024-SSA3-2012: numeral de información clínica básica del paciente.

## Estado de implementación

  Implementado en la misma sesión (2026-08-19).
  Función `ValidatePatientDemographics` en `medical_basic.go`.
  Handlers `HandlePatientCreate` y `HandlePatientUpdate` llaman la función.
  Issue #162 cerrado.

## Consecuencias

  Cualquier nueva ruta de ingreso de datos de paciente que llame
  `ValidatePatientDemographics` hereda automáticamente estas reglas sin
  duplicar código. La lógica clínica y la lógica HTTP están desacopladas.
  El Shader es el único lugar donde buscar qué se considera un dato
  demográfico válido.
