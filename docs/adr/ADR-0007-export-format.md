# ADR-0007 — Formato de export clínico: XML + JSON

## Estado
Propuesto

## Fecha
2026-06-22

## Contexto

El export actual de evidencia clínica genera un JSON mínimo que no
incluye el contenido clínico completo, no tiene mecanismo de
integridad verificable, y no sigue ningún estándar de
interoperabilidad.

Para que VUHMÍK sea defendible ante COFEPRIS, útil en auditorías
y compatible con el ecosistema de salud mexicano, el formato de
export debe evolucionar.

La NOM-024-SSA3-2012 exige formatos estructurados para ECE
interoperables. Los estándares internacionales de referencia son
HL7 CDA (XML) y FHIR (JSON/XML).

## Decisión

El export clínico de VUHMÍK soportará dos formatos:

### Formato 1 — JSON (uso interno, API, integraciones técnicas)

Estructura mínima obligatoria:

  {
    "format":       "vuhmik-evidence-v1",
    "evidence_id":  string,
    "tenant_id":    string,
    "subject_id":   string,
    "notes":        string,
    "state":        string,
    "created_at":   ISO8601,
    "issued_at":    ISO8601,
    "voided_at":    ISO8601 | null,
    "replaced_by_id": string | null,
    "hash":         "sha256:<hex>",
    "exported_at":  ISO8601
  }

El campo "hash" es SHA-256 del contenido canonico del documento
(todos los campos excepto "hash" y "exported_at"), serializado
en orden alfabético de claves, sin espacios.

### Formato 2 — XML (legal, auditoría, COFEPRIS, traspaso)

Estructura basada en HL7 CDA simplificado, adaptada al contexto
mexicano y a los requisitos mínimos de NOM-004-SSA3-2012:

  <?xml version="1.0" encoding="UTF-8"?>
  <ClinicalDocument xmlns="urn:vuhmik:hl7:v1">
    <id root="evidence_id"/>
    <effectiveTime value="issued_at"/>
    <confidentialityCode code="N"/>
    <recordTarget>
      <patientRole>
        <id root="subject_id"/>
      </patientRole>
    </recordTarget>
    <author>
      <time value="created_at"/>
      <assignedAuthor>
        <id root="tenant_id"/>
      </assignedAuthor>
    </author>
    <component>
      <structuredBody>
        <component>
          <section>
            <title>Nota clinica</title>
            <text>notes</text>
          </section>
        </component>
      </structuredBody>
    </component>
    <integrity>
      <hash algorithm="SHA-256">hex</hash>
      <state>issued</state>
      <replaced_by_id>string | null</replaced_by_id>
    </integrity>
  </ClinicalDocument>

El hash del documento XML se calcula sobre el contenido del
elemento <structuredBody> serializado en UTF-8, excluyendo el
elemento <integrity>.

## Rutas de export

  POST /api/v1/evidence/:id/export
    Header: Accept: application/json  → retorna JSON
    Header: Accept: application/xml   → retorna XML
    Default: application/json

  POST /api/v1/patients/:id/export
    Exporta el expediente completo del paciente:
    datos del paciente + todas las notas activas.
    Formatos: JSON o XML segun header Accept.

## Garantías

  - El archivo exportado NO se persiste en servidor.
  - Cache-Control: no-store en toda respuesta de export.
  - El hash permite verificar integridad sin acceso al servidor.
  - El formato XML es el oficial para auditorías COFEPRIS.
  - El formato JSON es el oficial para integraciones técnicas.

## Alternativas consideradas

  FHIR R4 completo — rechazado para v1 por complejidad.
  Solo JSON — rechazado: no es suficiente para auditoría legal.
  Solo XML — rechazado: no es práctico para API REST.

## Consecuencias

  Se requiere implementar generador de hash SHA-256 en Go.
  Se requiere generador de XML en Go (encoding/xml).
  El LegalExportShader debe extenderse para soportar ambos formatos.
  Requiere issue de implementación separado.
