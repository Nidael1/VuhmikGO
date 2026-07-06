package shaders

import (
	"encoding/json"
	"fmt"
	"time"
)

// IPSMedicationStatement representa el recurso MedicationStatement del IPS
// (ADR-0010, ADR-0011). Usado para la sección Medication Summary del IPS.
// Perfil FHIR R4 simplificado — texto libre en v1 (ADR-0011 §Decision).
type IPSMedicationStatement struct {
	ResourceType     string             `json:"resourceType"`
	ID               string             `json:"id"`
	Status           string             `json:"status"`
	MedicationCodeableConcept IPSCodeableConcept `json:"medicationCodeableConcept"`
	Subject          IPSReference       `json:"subject"`
	EffectiveDateTime string            `json:"effectiveDateTime,omitempty"`
	Dosage           []IPSDosage        `json:"dosage,omitempty"`
	Note             []IPSAnnotation    `json:"note,omitempty"`
}

// IPSDosage representa las instrucciones de dosificación FHIR R4.
type IPSDosage struct {
	Text string `json:"text,omitempty"`
}

// ProjectPrescriptionToIPS proyecta un registro de receta del Core
// al recurso FHIR MedicationStatement (ADR-0010, ADR-0011).
func ProjectPrescriptionToIPS(data ExportData) (*IPSMedicationStatement, error) {
	var content PrescriptionContent
	if err := json.Unmarshal([]byte(data.Content), &content); err != nil {
		return nil, fmt.Errorf("error al parsear contenido de receta: %w", err)
	}
	if content.Type != "prescription" {
		return nil, fmt.Errorf("tipo de contenido incorrecto: %s", content.Type)
	}

	effective := ""
	if data.IssuedAt != nil {
		effective = data.IssuedAt.Format(time.RFC3339)
	} else {
		effective = data.CreatedAt.Format(time.RFC3339)
	}

	// Texto de dosificación combinando dosis e indicaciones
	dosageText := content.Dosis
	if content.Indicaciones != "" {
		dosageText = content.Dosis + " — " + content.Indicaciones
	}

	med := &IPSMedicationStatement{
		ResourceType: "MedicationStatement",
		ID:           data.EvidenceID,
		Status:       "active",
		MedicationCodeableConcept: IPSCodeableConcept{
			Text: content.MedicamentoGenerico,
		},
		Subject:           IPSReference{Reference: "Patient/" + data.SubjectRef},
		EffectiveDateTime: effective,
	}

	if dosageText != "" {
		med.Dosage = []IPSDosage{{Text: dosageText}}
	}

	// Diagnóstico y seguimiento como notas (texto libre en v1)
	var notes []IPSAnnotation
	if content.Diagnostico != "" {
		notes = append(notes, IPSAnnotation{Text: "Diagnóstico: " + content.Diagnostico})
	}
	if content.Seguimiento != "" {
		notes = append(notes, IPSAnnotation{Text: "Seguimiento: " + content.Seguimiento})
	}
	if content.EsControlado {
		notes = append(notes, IPSAnnotation{Text: "ADVERTENCIA: Medicamento controlado — requiere recetario especial (COFEPRIS)"})
	}
	if len(notes) > 0 {
		med.Note = notes
	}

	return med, nil
}

// ExportPrescriptionAsIPS serializa una receta como JSON IPS MedicationStatement.
func ExportPrescriptionAsIPS(data ExportData) ([]byte, error) {
	med, err := ProjectPrescriptionToIPS(data)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(med, "", "  ")
}
