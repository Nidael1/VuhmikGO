package shaders

import (
	"encoding/json"
	"fmt"
	"time"
)

// IPSObservation representa el recurso Observation del IPS (ADR-0010, ADR-0015).
// Usado para resultados de laboratorio (sección Results del IPS).
type IPSObservation struct {
	ResourceType    string             `json:"resourceType"`
	ID              string             `json:"id"`
	Status          string             `json:"status"`
	Category        []IPSCodeableConcept `json:"category"`
	Code            IPSCodeableConcept `json:"code"`
	Subject         IPSReference       `json:"subject"`
	EffectiveDateTime string           `json:"effectiveDateTime,omitempty"`
	ValueString     string             `json:"valueString,omitempty"`
	Note            []IPSAnnotation    `json:"note,omitempty"`
}

// ProjectLabResultToIPS proyecta un resultado de laboratorio al recurso FHIR Observation.
func ProjectLabResultToIPS(data ExportData) (*IPSObservation, error) {
	var content LabResultContent
	if err := json.Unmarshal([]byte(data.Content), &content); err != nil {
		return nil, fmt.Errorf("error al parsear contenido de laboratorio: %w", err)
	}
	if content.Type != "lab_result" {
		return nil, fmt.Errorf("tipo de contenido incorrecto: %s", content.Type)
	}

	effective := content.FechaEstudio
	if effective == "" && data.IssuedAt != nil {
		effective = data.IssuedAt.Format(time.RFC3339)
	}

	valueStr := content.Resultado
	if content.Unidades != "" {
		valueStr = content.Resultado + " " + content.Unidades
	}

	obs := &IPSObservation{
		ResourceType: "Observation",
		ID:           data.EvidenceID,
		Status:       "final",
		Category: []IPSCodeableConcept{{
			Coding: []IPSCoding{{
				System:  "http://terminology.hl7.org/CodeSystem/observation-category",
				Code:    "laboratory",
				Display: "Laboratory",
			}},
		}},
		Code:              IPSCodeableConcept{Text: content.Estudio},
		Subject:           IPSReference{Reference: "Patient/" + data.SubjectRef},
		EffectiveDateTime: effective,
		ValueString:       valueStr,
	}
	if content.Notas != "" {
		obs.Note = []IPSAnnotation{{Text: content.Notas}}
	}
	return obs, nil
}

// ExportLabResultAsIPS serializa un resultado de laboratorio como JSON IPS Observation.
func ExportLabResultAsIPS(data ExportData) ([]byte, error) {
	obs, err := ProjectLabResultToIPS(data)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(obs, "", "  ")
}
