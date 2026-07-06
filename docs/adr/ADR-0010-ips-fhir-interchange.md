# ADR-0010 — Adoptar IPS sobre FHIR como modelo de intercambio y contenido clinico

## Estado
Aceptado

## Fecha
2026-06-24

## Contexto

VUHMIK exporta hoy la evidencia en JSON tecnico y en un XML tipo
HL7 CDA simplificado hecho a medida (ADR-0007), firmado con hash
SHA-256 (ADR-0008). El XML actual es un esquema propio: solo se
interpreta de forma garantizada entre instancias de VUHMIK.

La vision del producto es que un medico exporte el expediente y
otro medico lo reciba (exportador / receptor), compartiendo via
XML o JSON. Un estudio de mercado sobre los expedientes clinicos
de catorce paises (SS-MIX2 Japon, NHI Taiwan, MyHealthWay Corea,
RNDS Brasil, e-Nabiz Turquia, SATUSEHAT Indonesia, NEHR Singapur,
Malasia, ABDM India, Tailandia, Vietnam, Filipinas, Oman y el
Servicio Universal de Salud de Mexico) mostro que ese patron
corresponde al International Patient Summary (IPS): un estandar de
datos minimo, sobre HL7 FHIR R4, para compartir informacion clinica
esencial entre instituciones y fronteras.

Corea, India, Indonesia, Malasia y Oman ya adoptaron el IPS; los
tres ultimos fueron reconocidos en la Asamblea Mundial de la Salud
2024. En Mexico, la NOM-024-SSA3-2012 exige interoperabilidad con
HL7 / FHIR / DICOM y certificacion CENETEC, y el Servicio Universal
de Salud avanza hacia interoperabilidad FHIR en el sector publico.

Un esquema propio dejaria a VUHMIK aislado de esa red: solo
VUHMIK-a-VUHMIK se entenderia.

## Decision

### Modelo adoptado

VUHMIK adopta el International Patient Summary (IPS) sobre FHIR R4
como modelo canonico de:

  1. El contenido clinico estructurado del expediente.
  2. El intercambio entre prestadores (exportador / receptor).

El IPS se serializa indistintamente en JSON o en XML (FHIR soporta
ambos), por lo que la capacidad de exportar en ambos formatos
(ADR-0007) se conserva. Lo que cambia es el esquema: del XML CDA
propio al perfil IPS estandar.

### Restriccion arquitectonica

El IPS vive en la frontera de Shaders y en la capa de export.
NO vive en el Core.

  - El Core permanece agnostico: no conoce FHIR, ni IPS, ni
    reglas clinicas. Sigue almacenando registros append-only
    genericos.
  - El Shader proyecta los registros del Core a un documento IPS
    al exportar, y los valida al recibir.
  - El Core solo ve Create. No conoce el concepto de IPS, igual
    que no conoce el concepto de traspaso (ADR-0009).

### Integridad

El hash SHA-256 (ADR-0008) se aplica al documento IPS generado.

El documento compartido es asi un documento clinico verificable,
no un volcado de datos. Esto coincide con el hallazgo del estudio:
en Corea, los expertos definieron el IPS no como un paquete de
datos sino como un documento clinico creible, con fuentes
identificadas y firma.

### Secciones del IPS como roadmap

El IPS define las secciones que estructuran el expediente
compartido. Las funciones clinicas a agregar dejan de ser features
sueltas y se vuelven secciones del IPS:

  Obligatorias:
    - Lista de problemas / diagnosticos   (ADR-0013)
    - Alergias e intolerancias            (ADR-0012)
    - Resumen de medicacion / receta      (ADR-0011)

  Recomendadas:
    - Inmunizaciones / vacunacion         (ADR-0014)
    - Resultados de laboratorio           (ADR-0015)

## Dependencias

  - ADR-0007: formato de export XML+JSON (su esquema CDA propio
    queda como legado a deprecar, no a eliminar de inmediato).
  - ADR-0008: hash de integridad, ahora aplicado al documento IPS.
  - ADR-0009: protocolo de traspaso, que pasa a usar IPS como
    formato del archivo exportado/importado.

## Estado de implementacion

  Implementado. ips_bundle.go (IPS Bundle FHIR R4), legal_export.go y
  legal_export_xml.go actualizados al formato IPS canónico. Proyectores IPS
  para alergias, diagnósticos, inmunizaciones y laboratorio. Issue #212.
    - Incorporacion de perfiles FHIR R4 / IPS y su tooling.
    - Proyector de registros del Core a documento IPS (Shader).
    - Validador de documento IPS al recibir (Shader).
    - Terminologias estandarizadas (p. ej. CIE-10) en las
      secciones que lo exijan.
    - Migracion progresiva del export CDA propio al perfil IPS.

## Consecuencias

  El receptor lee el expediente aunque use otro sistema.
  Compatibilidad con la plataforma nacional y con CENETEC/NOM-024.
  Las funciones clinicas a agregar quedan definidas como secciones
  IPS con estructura y codificacion estandarizadas.
  Refuerza la tesis de producto: el documento compartido es creible
  y firmado, no un volcado de datos.
  El esquema CDA propio de ADR-0007 queda como legado a deprecar.
  Mayor complejidad que el XML propio (perfiles FHIR, terminologias).
