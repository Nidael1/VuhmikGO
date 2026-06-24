# ADR-0015 — Modulo de resultados de laboratorio

## Estado
Propuesto

## Fecha
2026-06-24

## Contexto

La investigacion de mercado (14 paises, junio 2026) identifico los
resultados de laboratorio como el segundo minimo funcional de la
NOM-024-SSA3-2012 que VUHMIK no cubre (junto con la receta,
ADR-0011). Japon (SS-MIX2), Taiwan (NHI MediCloud), Corea
(MyHealthWay) e Indonesia (SATUSEHAT) lo incluyen todos como
dato central del expediente.

Es la segunda de las cuatro secciones recomendadas del IPS
(ADR-0010): Results (laboratorio, patologia, radiologia).

El valor clinico para el medico independiente es inmediato:
  - Vincular el resultado al problema que lo origino (ADR-0013).
  - Tener el historial de laboratorio en el expediente exportado
    al recibir un paciente trasladado (ADR-0009).
  - Evitar la duplicacion de estudios, que fue el beneficio
    central del sistema taiwanes (ahorro de 8% en laboratorio).

En VUHMIK v1 el medico puede escribir resultados dentro de una
nota clinica (texto libre), pero no existe un registro estructurado,
vinculado al paciente, que pueda proyectarse al IPS.

El alcance de v1 cubre laboratorio clinico (biometria hematica,
quimica sanguinea, orina, cultivos, etc.). Radiologia e imagenes
(DICOM) quedan fuera del alcance por complejidad y peso, y se
documentaran en un ADR futuro.

## Decision

### Modelo de datos

Un resultado de laboratorio es un registro append-only en el Core.
El Core no conoce el concepto de "resultado" — solo ve un registro
con SubjectID (ID del paciente), contenido estructurado y estado.

El Shader MedicalBasicShader es la unica via de acceso al Core
para crear un resultado de laboratorio.

### Lifecycle

Los resultados siguen el mismo lifecycle que la evidencia:

  draft -> issued -> locked

  - draft:  el medico carga el resultado; no es parte del
            expediente activo.
  - issued: el resultado es parte del expediente del paciente.
  - locked: el resultado fue archivado; sigue visible en el
            historial pero no en la vista activa.

Correcciones: solo mediante void + replace (ADR-0006).
No existe edicion ni borrado de resultados registrados.

### Campos del registro

Obligatorios para emitir:

  - patient_id:      referencia al paciente en el mismo tenant.
  - estudio:         nombre del estudio (texto libre en v1).
  - fecha_estudio:   fecha en que se realizo el estudio.
  - resultado:       texto del resultado o valores encontrados.

Opcionales:

  - laboratorio:     nombre del laboratorio que realizo el estudio.
  - solicitado_por:  medico que solicito el estudio.
  - diagnostico_ref: referencia al diagnostico que origino el
                     estudio (patient_id + diagnose_id).
  - valores_ref:     rango de referencia del laboratorio.
  - interpretacion:  normal / anormal / critico.
  - notas:           observaciones del medico sobre el resultado.
  - archivo_ref:     referencia externa al PDF del resultado
                     (URL o nombre de archivo); el archivo NO
                     se almacena en VUHMIK en v1.

El campo "estudio" es texto libre en v1. En v2 se codificara
con LOINC para cumplir el perfil IPS completo.

### Almacenamiento de archivos

VUHMIK v1 NO almacena archivos PDF de resultados.

Razon: el almacenamiento de archivos binarios implica decisiones
de infraestructura (S3, volumen, cifrado en reposo de archivos)
que van mas alla del alcance actual y requieren su propio ADR.

El medico puede registrar una referencia externa (URL, nombre
de archivo) en el campo archivo_ref, pero el archivo vive fuera
de VUHMIK. Esta decision se revisara en v2.

### Seccion IPS

El resultado se modela como Observation (Results: laboratory)
del IPS (ADR-0010).

El Shader LegalExportShader proyecta los registros de laboratorio
del Core al perfil IPS al exportar.

Radiologia (Observation Results: radiology) e imagenes (ImagingStudy)
quedan fuera del alcance de este ADR.

### Integracion con diagnosticos

Un resultado puede referenciar al diagnostico que lo origino
(diagnostico_ref). Esta referencia es opcional y no implica
ninguna logica en el Core — es solo metadata que el Shader
puede incluir en el export IPS.

### Integracion con el modulo de pacientes

Cada resultado esta vinculado obligatoriamente a un paciente
(patient_id en el mismo tenant). No existe resultado sin paciente.

El historial de laboratorio es visible en el detalle del paciente,
ordenado por fecha del estudio.

## Dependencias

  - ADR-0006: void + replace como unica via de correccion.
  - ADR-0008: hash SHA-256 aplicado al registro al emitir.
  - ADR-0009: los resultados son parte del expediente exportado
              en el traspaso de paciente.
  - ADR-0010: los resultados son la seccion Results del IPS
              (recomendada).
  - ADR-0013: un resultado puede referenciar al diagnostico
              que lo origino.

## Estado de implementacion

  No implementado en v1.
  Requiere issues de implementacion con:
    - Migracion: tabla lab_results (id, tenant_id, patient_id,
      estudio, fecha_estudio, resultado, laboratorio,
      solicitado_por, diagnostico_ref, valores_ref,
      interpretacion, notas, archivo_ref, estado, hash,
      created_at, issued_at, voided_at, replaced_by_id).
    - Shader: validacion de campos minimos antes de emitir.
    - Handler API: POST /api/v1/lab-results/draft,
      POST /api/v1/lab-results/:id/emit,
      POST /api/v1/lab-results/:id/void.
    - Handler API: GET /api/v1/patients/:id/lab-results.
    - Export IPS: proyeccion a Observation (Results: laboratory).
    - Frontend: seccion de laboratorio en detalle del paciente,
      ordenada por fecha del estudio.
    - Frontend: campo de referencia externa para PDF del resultado.

## Consecuencias

  El medico registra resultados de laboratorio estructurados.
  El expediente exportado incluye historial de laboratorio.
  Los resultados son inmutables, firmados y trazables.
  VUHMIK cumple la seccion recomendada de resultados del IPS
  y el minimo funcional de la NOM-024.
  El texto libre en v1 permite adopcion rapida; la codificacion
  LOINC queda para v2.
  El almacenamiento de archivos PDF queda fuera de v1, reduciendo
  la complejidad de infraestructura.
  Radiologia e imagenes DICOM quedan fuera del alcance de este ADR.
