package shaders

import (
	"encoding/json"
	"fmt"
	"time"
)

// IPSImmunization representa el recurso Immunization del IPS (ADR-0010, ADR-0014).
type IPSImmunization struct {
	ResourceType       string             `json:"resourceType"`
	ID                 string             `json:"id"`
	Status             string             `json:"status"`
	VaccineCode        IPSCodeableConcept `json:"vaccineCode"`
	Patient            IPSReference       `json:"patient"`
	OccurrenceDateTime string             `json:"occurrenceDateTime,omitempty"`
	LotNumber          string             `json:"lotNumber,omitempty"`
	Note               []IPSAnnotation    `json:"note,omitempty"`
}

// IPSAnnotation es una nota de texto libre en FHIR.
type IPSAnnotation struct {
	Text string `json:"text"`
}

// ProjectImmunizationToIPS proyecta un registro de vacuna al recurso FHIR Immunization.
func ProjectImmunizationToIPS(data ExportData) (*IPSImmunization, error) {
	var content ImmunizationContent
	if err := json.Unmarshal([]byte(data.Content), &content); err != nil {
		return nil, fmt.Errorf("error al parsear contenido de vacuna: %w", err)
	}
	if content.Type != "immunization" {
		return nil, fmt.Errorf("tipo de contenido incorrecto: %s", content.Type)
	}

	occurrence := content.FechaAplicacion
	if occurrence == "" && data.IssuedAt != nil {
		occurrence = data.IssuedAt.Format(time.RFC3339)
	}

	imm := &IPSImmunization{
		ResourceType: "Immunization",
		ID:           data.EvidenceID,
		Status:       "completed",
		VaccineCode:  IPSCodeableConcept{Text: content.Vacuna},
		Patient:      IPSReference{Reference: "Patient/" + data.SubjectRef},
		OccurrenceDateTime: occurrence,
		LotNumber:    content.Lote,
	}
	if content.Notas != "" {
		imm.Note = []IPSAnnotation{{Text: content.Notas}}
	}
	return imm, nil
}

// ExportImmunizationAsIPS serializa una vacuna como JSON IPS Immunization.
func ExportImmunizationAsIPS(data ExportData) ([]byte, error) {
	imm, err := ProjectImmunizationToIPS(data)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(imm, "", "  ")
}
