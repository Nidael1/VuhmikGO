package api

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Nidael1/VuhmikGO/internal/core/evidence"
)

type fhirBundle struct {
	ResourceType string      `json:"resourceType"`
	Type         string      `json:"type"`
	Meta         fhirMeta    `json:"meta"`
	Entry        []fhirEntry `json:"entry"`
}

type fhirMeta struct {
	Source string `json:"source"`
}

type fhirEntry struct {
	FullURL  string                 `json:"fullUrl"`
	Resource map[string]interface{} `json:"resource"`
}

func importSource(bundle fhirBundle) string {
	src := strings.ToLower(bundle.Meta.Source)
	if strings.Contains(src, "imss") { return "fhir-imss" }
	if strings.Contains(src, "issste") { return "fhir-issste" }
	return "fhir-external"
}

func parseIPSBundleExternal(bundle fhirBundle, patientID, tenantID, source string) ([]evidence.Evidence, []string, error) {
	records := []evidence.Evidence{}
	warnings := []string{}
	now := time.Now().UTC()

	for _, entry := range bundle.Entry {
		rt, _ := entry.Resource["resourceType"].(string)
		if rt == "" { continue }

		var blob map[string]interface{}
		var evType string

		switch rt {
		case "AllergyIntolerance":
			blob, evType = parseAllergyIntolerance(entry.Resource)
		case "Condition":
			blob, evType = parseCondition(entry.Resource)
		case "Immunization":
			blob, evType = parseImmunization(entry.Resource)
		case "Observation":
			blob, evType = parseObservation(entry.Resource)
		case "MedicationStatement", "MedicationRequest":
			blob, evType = parseMedication(entry.Resource)
		case "Composition":
			blob, evType = parseComposition(entry.Resource)
		case "Patient", "Organization", "Practitioner":
			continue
		default:
			blob = map[string]interface{}{
				"type": "fhir_unknown",
				"resource_type": rt,
				"raw": entry.Resource,
				"import_source": source,
			}
			evType = "fhir_unknown"
			warnings = append(warnings, fmt.Sprintf("recurso %s no reconocido guardado como fhir_unknown", rt))
		}

		if blob == nil { continue }
		blob["import_source"] = source
		blob["import_ref"] = entry.FullURL
		blob["hash"] = "N/A-external"

		content, err := json.Marshal(blob)
		if err != nil { continue }

		id := "ev-fhir-" + now.Format("20060102150405.000") + "-" + strings.ReplaceAll(evType, "_", "-")
		e := evidence.Evidence{
			ID: id, TenantID: tenantID, SubjectRef: patientID,
			Content: string(content), State: evidence.StateIssued,
			CreatedAt: now, IssuedAt: &now,
		}
		records = append(records, e)
	}
	return records, warnings, nil
}

func parseAllergyIntolerance(r map[string]interface{}) (map[string]interface{}, string) {
	return map[string]interface{}{
		"type": "allergy",
		"agente": codeText(r, "code"),
		"tipo_reaccion": strField(r, "type"),
		"criticidad": strField(r, "criticality"),
	}, "allergy"
}

func parseCondition(r map[string]interface{}) (map[string]interface{}, string) {
	return map[string]interface{}{
		"type": "diagnosis",
		"descripcion": codeText(r, "code"),
		"codigo_cie10": codingCode(r, "code", "http://hl7.org/fhir/sid/icd-10"),
		"estado_problema": clinicalStatus(r),
		"fecha_inicio": strField(r, "onsetDateTime"),
	}, "diagnosis"
}

func parseImmunization(r map[string]interface{}) (map[string]interface{}, string) {
	return map[string]interface{}{
		"type": "immunization",
		"vacuna": codeText(r, "vaccineCode"),
		"fecha_aplicacion": strField(r, "occurrenceDateTime"),
		"lote": strField(r, "lotNumber"),
	}, "immunization"
}

func parseObservation(r map[string]interface{}) (map[string]interface{}, string) {
	return map[string]interface{}{
		"type": "lab_result",
		"estudio": codeText(r, "code"),
		"fecha_estudio": strField(r, "effectiveDateTime"),
		"resultado": valueString(r),
	}, "lab_result"
}

func parseMedication(r map[string]interface{}) (map[string]interface{}, string) {
	med := ""
	if mc, ok := r["medicationCodeableConcept"].(map[string]interface{}); ok {
		if t, ok := mc["text"].(string); ok { med = t }
	}
	dosis := ""
	if dosages, ok := r["dosage"].([]interface{}); ok && len(dosages) > 0 {
		if d, ok := dosages[0].(map[string]interface{}); ok { dosis, _ = d["text"].(string) }
	}
	return map[string]interface{}{
		"type": "prescription",
		"medicamento_generico": med,
		"dosis": dosis,
		"cedula_profesional": "N/A-external",
		"especialidad": "N/A-external",
	}, "prescription"
}

func parseComposition(r map[string]interface{}) (map[string]interface{}, string) {
	title, _ := r["title"].(string)
	text := ""
	if sections, ok := r["section"].([]interface{}); ok && len(sections) > 0 {
		if s, ok := sections[0].(map[string]interface{}); ok {
			if t, ok := s["text"].(map[string]interface{}); ok { text, _ = t["div"].(string) }
		}
	}
	if title == "" && text == "" { return nil, "" }
	return map[string]interface{}{"type": "note", "text": title + " " + text}, "note"
}

func codeText(r map[string]interface{}, field string) string {
	var m map[string]interface{}
	if field == "" { m = r } else if v, ok := r[field].(map[string]interface{}); ok { m = v } else { return "" }
	if text, ok := m["text"].(string); ok { return text }
	if codings, ok := m["coding"].([]interface{}); ok && len(codings) > 0 {
		if c, ok := codings[0].(map[string]interface{}); ok {
			if d, ok := c["display"].(string); ok { return d }
		}
	}
	return ""
}

func codingCode(r map[string]interface{}, field, system string) string {
	if v, ok := r[field].(map[string]interface{}); ok {
		if codings, ok := v["coding"].([]interface{}); ok {
			for _, ci := range codings {
				if c, ok := ci.(map[string]interface{}); ok {
					if s, _ := c["system"].(string); s == system {
						if code, ok := c["code"].(string); ok { return code }
					}
				}
			}
		}
	}
	return ""
}

func strField(r map[string]interface{}, field string) string {
	if v, ok := r[field].(string); ok { return v }
	return ""
}

func clinicalStatus(r map[string]interface{}) string {
	if cs, ok := r["clinicalStatus"].(map[string]interface{}); ok { return codeText(cs, "") }
	return ""
}

func valueString(r map[string]interface{}) string {
	if v, ok := r["valueString"].(string); ok { return v }
	if v, ok := r["valueQuantity"].(map[string]interface{}); ok {
		val, _ := v["value"].(float64)
		unit, _ := v["unit"].(string)
		return fmt.Sprintf("%.2f %s", val, unit)
	}
	return ""
}
