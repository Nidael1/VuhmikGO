# ADR-0014 — Modulo de inmunizaciones y vacunacion

## Estado
Propuesto

## Fecha
2026-06-24

## Contexto

La investigacion de mercado (14 paises, junio 2026) identifico
la vacunacion como una funcion de alto valor y bajo costo de
implementacion. Taiwan (NHI MediCloud) la integra directamente
desde el CDC nacional; India la vuelve obligatoria en su perfil
IPS nacional (IN PS); Indonesia (SATUSEHAT) la incluye como
dato de encuentro desde el lanzamiento.

Es la primera de las cuatro secciones recomendadas del IPS
(ADR-0010): Immunizations. Su valor clinico es claro para el
medico general independiente: es uno de los datos que mas
frecuentemente se desconoce cuando se recibe un paciente nuevo
o trasladado, y uno de los que mas impacta en la prevencion.

En Mexico la Cartilla Nacional de Vacunacion es el documento
oficial de registro. El Sistema Nacional de Vigilancia
Epidemiologica (SINAVE) centraliza los datos de vacunacion del
sector publico, pero el medico privado no tiene acceso a esos
datos ni obligacion de reportarlos en v1. El medico privado
registra las vacunas que el mismo aplica o que el paciente
reporta haber recibido.

## Decision

### Modelo de datos

Un registro de vacuna es un registro append-only en el Core.
El Core no conoce el concepto de "vacuna" — solo ve un registro
con SubjectID (ID del paciente), contenido estructurado y estado.

El Shader MedicalBasicShader es la unica via de acceso al Core
para crear un registro de vacuna.

### Lifecycle

Las vacunas siguen el mismo lifecycle que la evidencia:

  draft -> issued -> locked

  - draft:  el medico registra la vacuna; no es activa.
  - issued: la vacuna es parte del historial del paciente.
  - locked: no aplica en el uso tipico; reservado para
            registros que requieran congelamiento explicito.

Correcciones: solo mediante void + replace (ADR-0006).
No existe edicion ni borrado de registros de vacuna.

### Campos del registro

Obligatorios para emitir:

  - patient_id:      referencia al paciente en el mismo tenant.
  - vacuna:          nombre de la vacuna (texto libre en v1).
  - fecha_aplicacion: fecha en que se aplico o reporto.

Opcionales:

  - lote:            numero de lote del bilogico.
  - dosis:           numero de dosis (primera, segunda, refuerzo).
  - via:             via de administracion (intramuscular, oral, etc).
  - aplicada_por:    quien aplico (medico actual, otro, reportada
                     por el paciente).
  - notas:           observaciones adicionales.

El campo "vacuna" es texto libre en v1. En v2 se codificara con
terminologia estandar (CVX o SNOMED CT) para cumplir el perfil
IPS completo.

### Fuentes del registro

El medico puede registrar dos tipos de vacuna:

  Aplicada en consulta: el medico la aplico directamente.
  Reportada por el paciente: el paciente informa haber recibido
  la vacuna en otro establecimiento (sector publico, otro medico).

Ambas son validas y se distinguen por el campo "aplicada_por".
El sistema no verifica contra el SINAVE en v1.

### Seccion IPS

La vacuna se modela como Immunization del IPS (ADR-0010).
El Shader LegalExportShader proyecta los registros de vacuna
del Core al perfil IPS al exportar.

### Integracion con el modulo de pacientes

Cada vacuna esta vinculada obligatoriamente a un paciente
(patient_id en el mismo tenant). No existe registro de vacuna
sin paciente.

El historial de vacunacion es visible en el detalle del paciente,
ordenado por fecha de aplicacion.

## Dependencias

  - ADR-0006: void + replace como unica via de correccion.
  - ADR-0008: hash SHA-256 aplicado al registro al emitir.
  - ADR-0009: el historial de vacunacion es parte del expediente
              exportado en el traspaso de paciente.
  - ADR-0010: las vacunas son la seccion Immunizations del IPS
              (recomendada).

## Estado de implementacion

  No implementado en v1.
  Requiere issues de implementacion con:
    - Migracion: tabla immunizations (id, tenant_id, patient_id,
      vacuna, fecha_aplicacion, lote, dosis, via, aplicada_por,
      notas, estado, hash, created_at, issued_at, voided_at,
      replaced_by_id).
    - Shader: validacion de campos minimos antes de emitir.
    - Handler API: POST /api/v1/immunizations/draft,
      POST /api/v1/immunizations/:id/emit,
      POST /api/v1/immunizations/:id/void.
    - Handler API: GET /api/v1/patients/:id/immunizations.
    - Export IPS: proyeccion a Immunization.
    - Frontend: seccion de vacunas en detalle del paciente,
      ordenada por fecha.

## Consecuencias

  El medico registra el historial de vacunacion del paciente.
  El expediente exportado incluye vacunas aplicadas y reportadas.
  Los registros son inmutables, firmados y trazables.
  VUHMIK cumple la seccion recomendada de inmunizaciones del IPS.
  El texto libre en v1 permite adopcion rapida; la codificacion
  CVX/SNOMED queda para v2.
  Sin integracion con SINAVE en v1 — el medico privado registra
  lo que aplica o lo que el paciente reporta.
