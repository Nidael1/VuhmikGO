# ADR-0011 — Modulo de medicacion y receta electronica

## Estado
Propuesto

## Fecha
2026-06-24

## Contexto

La investigacion de mercado (14 paises, junio 2026) identifico la
receta electronica como la funcion con mayor impacto inmediato para
el medico independiente en Mexico:

  - Es una de las tres secciones obligatorias del IPS (ADR-0010):
    Medication Summary.
  - Es uno de los cuatro minimos funcionales de la NOM-024-SSA3-2012:
    notas clinicas, prescripcion, resultados, seguridad.
  - Es la funcion que ya tienen todos los competidores directos
    (SaludTotal, Doctocliq, Davix, Nimbo-x).
  - El breach de SingHealth (2018) robo 160,000 registros de
    prescripcion, validando que es el dato de mayor valor clinico
    y mayor riesgo; debe ser append-only y firmado desde el dia uno.
  - La NOM-024 define los campos minimos de validez legal:
    cedula profesional, especialidad, datos del paciente, nombre
    generico del medicamento, dosis y firma del medico.

En Mexico existen dos categorias de medicamentos con reglas
distintas:

  Medicamentos de libre venta y de receta simple:
    La receta electronica es plenamente valida bajo NOM-024 y la
    Ley General de Salud (Art. 28 Bis y 226).

  Medicamentos controlados COFEPRIS (Grupo II y III):
    Requieren recetario especial de la Secretaria de Salud con
    codigo de barras y folio. La emision digital tiene reglas
    especificas que varian por estado y requieren validacion
    regulatoria explicita antes de implementarse.

## Decision

### Modelo de datos

Una receta es un registro append-only en el Core.
El Core no conoce el concepto de "receta" — solo ve un registro
con SubjectID (ID del paciente), contenido estructurado y estado.

El Shader MedicalBasicShader es la unica via de acceso al Core
para crear una receta. El Shader valida los campos minimos de
validez legal antes de permitir la emision.

### Lifecycle

La receta sigue el mismo lifecycle que la evidencia clinica:

  draft -> issued -> locked

  - draft:  el medico compone la receta; no tiene validez legal.
  - issued: el medico emite la receta; adquiere validez legal;
            es inmutable desde este momento.
  - locked: la receta fue dispensada o vencio; no se puede anular.

Correcciones: solo mediante void + replace (igual que ADR-0006).
No existe edicion ni borrado de recetas emitidas.

### Campos minimos de validez legal (NOM-024)

Obligatorios antes de poder emitir (issued):

  - cedula_profesional:  texto, no vacio.
  - especialidad:        texto, no vacio.
  - patient_id:          referencia al paciente en el mismo tenant.
  - medicamento_generico: nombre generico del medicamento, no vacio.
  - dosis:               texto con cantidad y frecuencia, no vacio.

El hash SHA-256 (ADR-0008) se calcula sobre la receta al emitirla.
La receta exportada como IPS Medication Summary incluye el hash.

### Medicamentos controlados

VUHMIK v1 NO implementa el flujo especial de medicamentos
controlados COFEPRIS (Grupo II y III).

Razon: las reglas de emision digital de controlados varian por
estado y requieren validacion regulatoria explicita con la
Jurisdiccion Sanitaria local antes de implementarse. Implementar
este flujo sin esa validacion expone al medico a riesgo legal.

El sistema debe mostrar una advertencia visible cuando el medico
intente prescribir un medicamento de la lista de controlados,
indicando que debe usar el recetario especial en papel.

La lista de controlados y el flujo digital se documentaran en un
ADR futuro una vez obtenida la validacion regulatoria.

### Sección IPS

La receta se modela como Medication Summary del IPS (ADR-0010):

  - MedicationStatement para medicacion activa o historica.
  - MedicationRequest para prescripciones nuevas.

El Shader LegalExportShader proyecta los registros de receta
del Core al perfil IPS al exportar.

### Integracion con el modulo de pacientes

Cada receta esta vinculada obligatoriamente a un paciente
(patient_id en el mismo tenant). No existe receta sin paciente.

## Dependencias

  - ADR-0006: void + replace como unica via de correccion.
  - ADR-0008: hash SHA-256 aplicado a la receta al emitir.
  - ADR-0009: la receta es parte del expediente exportado
              en el traspaso de paciente.
  - ADR-0010: la receta es el Medication Summary del IPS.

## Estado de implementacion

  No implementado en v1.
  Requiere issues de implementacion con:
    - Migracion: tabla prescriptions (id, tenant_id, patient_id,
      cedula_profesional, especialidad, medicamento_generico,
      dosis, estado, hash, created_at, issued_at, voided_at,
      replaced_by_id).
    - Shader: validacion de campos minimos antes de emitir.
    - Handler API: POST /api/v1/prescriptions/draft,
      POST /api/v1/prescriptions/:id/emit,
      POST /api/v1/prescriptions/:id/void.
    - Handler API: GET /api/v1/patients/:id/prescriptions.
    - Export IPS: proyeccion a MedicationStatement / MedicationRequest.
    - Frontend: formulario de receta vinculado al detalle del paciente.
    - Advertencia visible para medicamentos controlados.

## Consecuencias

  El medico puede emitir recetas con validez legal desde VUHMIK.
  El expediente exportado incluye el historial de medicacion.
  Las recetas son inmutables, firmadas y trazables (COFEPRIS).
  Los medicamentos controlados quedan excluidos hasta validacion
  regulatoria, protegiendo al medico de riesgo legal.
  VUHMIK cumple con el minimo funcional de prescripcion de la NOM-024.
