# ADR-0012 — Modulo de alergias e intolerancias

## Estado
Propuesto

## Fecha
2026-06-24

## Contexto

La investigacion de mercado (14 paises, junio 2026) identifico el
registro de alergias como una de las funciones mas criticas para
la seguridad del paciente y una de las mas baratas de implementar.
Japon (SS-MIX2), Taiwan (NHI MediCloud) y Corea (MyHealthWay) la
incluyen todos. Taiwan la lista entre sus 12 tipos de informacion
prioritarios; India la vuelve obligatoria en su perfil IPS nacional.

Es la segunda de las tres secciones obligatorias del IPS (ADR-0010):
Allergies and Intolerances.

Su valor clinico central: un medico que recibe a un paciente
trasladado (ADR-0009) o que va a prescribir (ADR-0011) necesita
conocer las alergias antes de cualquier otra accion. Es la
informacion de seguridad mas urgente en una consulta de primera vez.

El breach de SingHealth (2018) demostro que los datos de
prescripcion son los de mayor valor y riesgo; el registro de
alergias es su complemento inseparable: sin alergias, la receta
no puede considerarse segura.

## Decision

### Modelo de datos

Una alergia es un registro append-only en el Core.
El Core no conoce el concepto de "alergia" — solo ve un registro
con SubjectID (ID del paciente), contenido estructurado y estado.

El Shader MedicalBasicShader es la unica via de acceso al Core
para crear un registro de alergia.

### Lifecycle

Las alergias siguen el mismo lifecycle que la evidencia:

  draft -> issued -> locked

  - draft:  el medico registra la alergia; no es activa.
  - issued: la alergia es activa y visible en el expediente.
  - locked: la alergia fue resuelta o descartada; sigue visible
            en el historial pero marcada como inactiva.

Correcciones: solo mediante void + replace (ADR-0006).
No existe edicion ni borrado de alergias registradas.

Si una alergia desaparece o era incorrecta, se anula (void) y
se reemplaza con el registro correcto o con estado "resuelta".
El historial completo permanece siempre.

### Campos del registro

Obligatorios para emitir:

  - patient_id:      referencia al paciente en el mismo tenant.
  - agente:          sustancia causante (medicamento, alimento,
                     material, otro), texto libre en v1.
  - tipo_reaccion:   descripcion de la reaccion (rash, anafilaxia,
                     nausea, etc.), texto libre en v1.

Opcionales:

  - criticidad:      leve / moderada / grave.
  - certeza:         confirmada / sospecha / descartada.
  - fecha_inicio:    fecha de primera reaccion conocida.
  - notas:           observaciones adicionales del medico.

El campo "agente" es texto libre en v1. En v2 se codificara con
terminologia estandar (SNOMED CT o RxNorm) para cumplir el perfil
IPS completo.

### Visibilidad en el expediente

Las alergias activas (issued) son visibles de forma prominente
en la vista de detalle del paciente, antes del historial de notas.

Las alergias anuladas (voided) son visibles solo en el historial
completo, con indicacion de que fueron reemplazadas.

### Seccion IPS

La alergia se modela como AllergyIntolerance del IPS (ADR-0010).
El Shader LegalExportShader proyecta los registros de alergia
del Core al perfil IPS al exportar.

### Integracion con receta

Al emitir una receta (ADR-0011), el Shader debe verificar que
el medico haya revisado las alergias activas del paciente.
En v1 esta verificacion es informativa (muestra las alergias
activas al componer la receta). En v2 podra ser bloqueante para
agentes especificos.

### Integracion con el modulo de pacientes

Cada alergia esta vinculada obligatoriamente a un paciente
(patient_id en el mismo tenant). No existe alergia sin paciente.

## Dependencias

  - ADR-0006: void + replace como unica via de correccion.
  - ADR-0008: hash SHA-256 aplicado al registro al emitir.
  - ADR-0009: las alergias activas son parte del expediente
              exportado en el traspaso de paciente.
  - ADR-0010: las alergias son la seccion AllergyIntolerance
              del IPS (obligatoria).
  - ADR-0011: las alergias activas se muestran al componer
              una receta.

## Estado de implementacion

  No implementado en v1.
  Requiere issues de implementacion con:
    - Migracion: tabla allergies (id, tenant_id, patient_id,
      agente, tipo_reaccion, criticidad, certeza, fecha_inicio,
      notas, estado, hash, created_at, issued_at, voided_at,
      replaced_by_id).
    - Shader: validacion de campos minimos antes de emitir.
    - Handler API: POST /api/v1/allergies/draft,
      POST /api/v1/allergies/:id/emit,
      POST /api/v1/allergies/:id/void.
    - Handler API: GET /api/v1/patients/:id/allergies.
    - Export IPS: proyeccion a AllergyIntolerance.
    - Frontend: seccion de alergias en detalle del paciente,
      visible antes del historial de notas.
    - Frontend: alergias activas visibles al componer receta.

## Consecuencias

  El medico tiene acceso inmediato a las alergias del paciente
  antes de prescribir o al recibirlo por traspaso.
  El expediente exportado incluye alergias activas e historicas.
  Las alergias son inmutables, firmadas y trazables.
  VUHMIK cumple la seccion obligatoria de alergias del IPS.
  El agente en texto libre en v1 permite adopcion rapida;
  la codificacion SNOMED/RxNorm queda para v2.
