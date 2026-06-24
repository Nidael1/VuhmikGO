# ADR-0013 — Modulo de diagnosticos estructurados y lista de problemas (CIE-10)

## Estado
Propuesto

## Fecha
2026-06-24

## Contexto

La investigacion de mercado (14 paises, junio 2026) mostro que
Japon (SS-MIX2), Taiwan (NHI MediCloud), Corea (MyHealthWay) y
Brasil (RNDS) estructuran los diagnosticos con codificacion
estandar (ICD-10 / CIE-10). Es la tercera seccion obligatoria
del IPS (ADR-0010): Problem List.

En VUHMIK v1 las notas clinicas son texto libre. Un diagnostico
no es distinguible de una nota de evolucion o de cualquier otro
texto. Esto tiene dos consecuencias:

  1. El expediente exportado no puede incluir una lista de
     problemas estructurada — un campo central del IPS.
  2. La NOM-004-SSA3-2012 exige que el expediente incluya
     diagnosticos con interrogatorio por aparatos y sistemas.
     El texto libre no cumple esto de forma verificable.

La CIE-10 (Clasificacion Internacional de Enfermedades, decima
revision) es el estandar oficial en Mexico para codificacion
de diagnosticos, usado por IMSS, ISSSTE y el sector privado
con certificacion CENETEC.

## Decision

### Modelo de datos

Un diagnostico es un registro append-only en el Core.
El Core no conoce el concepto de "diagnostico" — solo ve un
registro con SubjectID (ID del paciente), contenido estructurado
y estado.

El Shader MedicalBasicShader es la unica via de acceso al Core
para crear un diagnostico.

### Lifecycle

Los diagnosticos siguen el mismo lifecycle que la evidencia:

  draft -> issued -> locked

  - draft:  el medico registra el diagnostico; no es activo.
  - issued: el diagnostico es activo en el expediente.
  - locked: el problema fue resuelto; sigue visible en el
            historial de problemas pasados.

Correcciones: solo mediante void + replace (ADR-0006).
No existe edicion ni borrado de diagnosticos registrados.

### Campos del registro

Obligatorios para emitir:

  - patient_id:        referencia al paciente en el mismo tenant.
  - descripcion:       texto libre del diagnostico clinico.

Opcionales en v1, obligatorios en v2:

  - codigo_cie10:      codigo CIE-10 (p. ej. "J06.9", "E11").
  - tipo:              principal / secundario / diferencial.
  - estado_problema:   activo / resuelto / cronico.
  - fecha_inicio:      fecha de inicio del problema.
  - notas:             observaciones adicionales.

El campo codigo_cie10 es opcional en v1 para permitir adopcion
rapida. El medico puede registrar el diagnostico en texto libre
y agregar el codigo despues mediante void + replace.

En v2 el codigo CIE-10 sera obligatorio para cumplir el perfil
IPS completo y la certificacion CENETEC.

### Catalogo CIE-10

En v1 el medico escribe el codigo manualmente (texto libre
validado contra formato basico CIE-10: letra + digitos).

En v2 se incorpora un catalogo de busqueda de codigos CIE-10
en el frontend, con busqueda por texto o por codigo.

### Seccion IPS

El diagnostico se modela como Condition del IPS (ADR-0010),
que es el recurso FHIR que representa tanto problemas activos
como historia de enfermedades pasadas.

El Shader LegalExportShader proyecta los registros de
diagnostico del Core al perfil IPS al exportar, separando:
  - Problemas activos  → IPS Problem List
  - Problemas resueltos → IPS History of Past Illness

### Relacion con notas clinicas

Un diagnostico NO reemplaza a una nota clinica (evidencia).
Son registros distintos y complementarios:

  Nota clinica (evidencia): narrativa de la consulta, subjetiva
  y objetiva, texto libre del medico.

  Diagnostico: clasificacion estructurada del problema clinico,
  codificable con CIE-10, vinculada al paciente.

Una consulta puede generar una nota Y uno o mas diagnosticos.

### Integracion con el modulo de pacientes

Cada diagnostico esta vinculado obligatoriamente a un paciente
(patient_id en el mismo tenant). No existe diagnostico sin
paciente.

### Integracion con receta

Al emitir una receta (ADR-0011), el sistema puede mostrar los
diagnosticos activos del paciente como contexto clinico.

## Dependencias

  - ADR-0006: void + replace como unica via de correccion.
  - ADR-0008: hash SHA-256 aplicado al registro al emitir.
  - ADR-0009: los diagnosticos activos son parte del expediente
              exportado en el traspaso de paciente.
  - ADR-0010: los diagnosticos son la seccion Problem List
              del IPS (obligatoria).
  - ADR-0011: los diagnosticos activos se muestran como contexto
              al componer una receta.

## Estado de implementacion

  No implementado en v1.
  Requiere issues de implementacion con:
    - Migracion: tabla diagnoses (id, tenant_id, patient_id,
      descripcion, codigo_cie10, tipo, estado_problema,
      fecha_inicio, notas, estado, hash, created_at, issued_at,
      voided_at, replaced_by_id).
    - Shader: validacion de campos minimos antes de emitir.
    - Handler API: POST /api/v1/diagnoses/draft,
      POST /api/v1/diagnoses/:id/emit,
      POST /api/v1/diagnoses/:id/void.
    - Handler API: GET /api/v1/patients/:id/diagnoses.
    - Export IPS: proyeccion a Condition (activos e historicos).
    - Frontend: seccion de diagnosticos en detalle del paciente.
    - Frontend: campo de codigo CIE-10 con validacion de formato.

## Consecuencias

  El expediente exportado incluye lista de problemas estructurada.
  Los diagnosticos son inmutables, firmados y trazables.
  VUHMIK cumple la seccion obligatoria de lista de problemas del IPS.
  El texto libre en v1 permite adopcion rapida; la codificacion
  CIE-10 obligatoria queda para v2 junto con la certificacion CENETEC.
  Las notas clinicas y los diagnosticos coexisten como registros
  complementarios; no se fusionan ni se reemplazan mutuamente.
