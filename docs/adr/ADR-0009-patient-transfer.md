# ADR-0009 — Protocolo de traspaso de paciente entre tenants

## Estado
Propuesto

## Fecha
2026-06-22

## Contexto

Un paciente puede cambiar de medico. El nuevo medico necesita
acceso al historial clinico del paciente para dar continuidad
al tratamiento. En los sistemas asiaticos de referencia
(SS-MIX2 Japon, NHI Taiwan, EMR Corea), el identificador unico
nacional (numero de seguro social, NHI number) permite la
trazabilidad del paciente entre instituciones.

En Mexico, el CURP es el identificador unico oficial de cada
ciudadano y es el equivalente al identificador nacional de salud.

VUHMIK es multi-tenant: cada medico es un tenant aislado.
Los registros de un tenant nunca son visibles para otro tenant
(regla absoluta de aislamiento).

El traspaso debe respetar esta regla sin violarla.

## Decision

### Modelo de traspaso

El traspaso NO comparte registros entre tenants.
El traspaso genera una copia de los registros en el tenant destino.
Los registros originales permanecen en el tenant origen (inmutables).

### Flujo propuesto

  1. Medico A (origen) exporta el expediente del paciente:
       POST /api/v1/patients/:id/export
       Formato: XML (ADR-0007) con hash de integridad (ADR-0008)

  2. El archivo exportado se entrega al paciente o al Medico B
     por canal externo (email, USB, plataforma).

  3. Medico B (destino) importa el expediente:
       POST /api/v1/patients/import
       Body: archivo XML del paso 1

  4. El sistema verifica el hash de integridad del archivo.

  5. El sistema crea un nuevo paciente en el tenant del Medico B
     con los datos del paciente importado.

  6. El sistema crea nuevas evidencias en estado "issued" en el
     tenant del Medico B, con referencia al origen:
       import_source: "vuhmik-transfer-v1"
       import_ref:    evidence_id original

  7. Los registros importados son inmutables desde el momento
     de importacion (ya llegan en estado issued).

### Identificador de traspaso

El CURP del paciente es el campo que permite identificar si un
paciente importado ya existe en el sistema del Medico B.

Si el CURP ya existe en el tenant destino:
  El sistema alerta al medico y pregunta si fusionar o crear nuevo.

Si el CURP no existe:
  Se crea el paciente normalmente.

### Garantias

  - Los registros originales del Medico A no se modifican.
  - Los registros importados en el Medico B son nuevos registros
    con sus propios IDs y timestamps.
  - El hash del archivo importado es verificable en cualquier momento.
  - El CURP es el identificador de continuidad asistencial.
  - El Core no conoce el concepto de traspaso — solo ve Create.

## Dependencias

  - ADR-0007: formato de export XML+JSON
  - ADR-0008: hash de integridad
  - CURP como campo obligatorio en tabla patients (migracion futura)

## Estado de implementacion

  No implementado en v1.
  Requiere issue de implementacion con:
    - Endpoint POST /api/v1/patients/import
    - Verificador de hash
    - Logica de fusion/creacion por CURP
    - Migracion para agregar CURP a patients

## Consecuencias

  El sistema es trazable entre medicos sin violar aislamiento.
  El paciente es dueno de su expediente — puede llevarlo consigo.
  El CURP como identificador nacional permite interoperabilidad futura.
  Compatible con la vision de estandar nacional tipo SS-MIX2.
