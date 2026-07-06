# ADR-0028 — Importación de IPS Bundle FHIR R4 externo

## Estado
Aceptado

## Fecha
2026-07-06

---

## Contexto

VUHMÍK ya genera IPS Bundle FHIR R4 estándar (ADR-0010, issue #212).
El importador de traspaso (issue #221) solo acepta el formato propio
`vuhmik-transfer-v1`. Sin embargo, los principales competidores en el
mercado mexicano (IMSS, ISSSTE, otros ECEs) generan expedientes en
IPS/FHIR R4 estándar.

Aceptar IPS externo tiene valor estratégico triple:
1. El médico puede recibir el historial completo de un paciente que
   viene del sector público, sin reingreso manual de datos.
2. VUHMÍK convierte datos externos en propios desde el momento de la
   importación — el médico asume la custodia clínica y legal.
3. La cadena de custodia registra el origen externo, permitiendo
   identificar irregularidades y atribuirlas al sistema emisor.

El ADR-0010 reservaba la integración con sistemas externos para v2.
Este ADR levanta esa restricción específicamente para la importación
pasiva de IPS Bundle FHIR R4, sin integración activa ni push hacia
sistemas externos.

---

## Decisión

### 1. Aceptar IPS Bundle FHIR R4 externo en el endpoint de importación

Se extiende `POST /api/v1/patients/import` para detectar y procesar
IPS Bundles FHIR R4 externos (`resourceType: "Bundle"`, `type: "document"`).

El formato se detecta automáticamente:
- Si el body tiene `format: "vuhmik-transfer-v1"` → flujo propio (ADR-0009).
- Si el body tiene `resourceType: "Bundle"` → flujo externo (este ADR).

### 2. Clasificación de origen — dos clases distinguibles

Todos los registros importados llevan `import_source` en el blob:

- `import_source: "vuhmik-transfer-v1"` → origen VUHMÍK verificado.
- `import_source: "fhir-external"` → origen externo no verificado.
- `import_source: "fhir-imss"` / `"fhir-issste"` → si el Bundle incluye
  identificador del emisor en `Bundle.meta.source` o `Organization`.

### 3. Hash de integridad para registros externos

Los Bundles externos no llevan `_vuhmik_hash`. Se marca:
Esto distingue visualmente en auditoría qué registros tienen
integridad verificable (propios) y cuáles son de confianza delegada
(externos, integridad garantizada por el sistema emisor).

### 4. Recursos FHIR conocidos → blob estructurado

Cada recurso del Bundle se proyecta al Content blob correspondiente:

| Recurso FHIR | Tipo de blob |
|---|---|
| `AllergyIntolerance` | `type: "allergy"` |
| `Condition` | `type: "diagnosis"` |
| `Immunization` | `type: "immunization"` |
| `Observation` (category: laboratory) | `type: "lab_result"` |
| `MedicationStatement` / `MedicationRequest` | `type: "prescription"` |
| `Composition` (narrativa) | `type: "note"` |

### 5. Recursos FHIR desconocidos → blob extra preservado

Todo recurso del Bundle que no tenga mapeo conocido se importa como:

```json
{
  "type": "fhir_unknown",
  "resource_type": "<FHIR resourceType>",
  "raw": { ... contenido completo del recurso ... },
  "import_source": "fhir-external"
}
```

No se pierden datos. El médico tiene acceso al contenido original.

### 6. Identificación del paciente — CURP obligatorio

El CURP del paciente ya existe en el perfil del médico receptor.
El sistema busca al paciente por CURP en el tenant destino.

- CURP encontrado → importar sobre el paciente existente.
- CURP no encontrado → rechazar con mensaje:
  "Paciente no encontrado — registra al paciente con su CURP antes
  de importar su expediente externo."

El CURP puede venir en el Bundle (`Patient.identifier` con system
`urn:oid:2.16.484.1.1`) o no venir. Si no viene en el Bundle,
el médico lo busca por nombre/fecha y confirma la identidad.

### 7. El Core no cambia

El parser vive en la capa Delivery/API. El Core solo ve operaciones
`Create` con evidencia en estado `issued`. El agnosticismo del Core
se preserva.

---

## Consecuencias

- VUHMÍK puede recibir expedientes del IMSS, ISSSTE y cualquier
  sistema que emita IPS/FHIR R4 estándar.
- Los registros externos son distinguibles de los propios en auditoría.
- Irregularidades en expedientes importados son atribuibles al emisor.
- El médico VUHMÍK asume custodia clínica desde el momento de importación.
- Este ADR NO autoriza push hacia sistemas externos.
- Este ADR NO autoriza integración activa con SINAVE, IMSS o ISSSTE.
- Este ADR NO modifica el Core ni los Shaders.

## Documentos impactados

- ADR-0010 (levanta restricción de v2 para importación pasiva).
- ADR-0009 (extiende el flujo de importación con formato externo).
