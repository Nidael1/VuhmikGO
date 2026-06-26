package shaders

import (
	"encoding/json"
	"fmt"
	"time"
)

// IPSAllergyIntolerance representa la seccion AllergyIntolerance del IPS.
// Perfil FHIR R4 simplificado — ADR-0010, ADR-0012.
// El Core nunca conoce esta estructura; vive exclusivamente en Shaders/export.
type IPSAllergyIntolerance struct {
	ResourceType    string          `json:"resourceType"`
	ID              string          `json:"id"`
	ClinicalStatus  IPSCodeableConcept `json:"clinicalStatus"`
	VerificationStatus IPSCodeableConcept `json:"verificationStatus,omitempty"`
	Type            string          `json:"type,omitempty"`
	Criticality     string          `json:"criticality,omitempty"`
	Code            IPSCodeableConcept `json:"code"`
	Patient         IPSReference    `json:"patient"`
	RecordedDate    string          `json:"recordedDate"`
	Reaction        []IPSReaction   `json:"reaction,omitempty"`
}

// IPSCodeableConcept representa un concepto codeable en FHIR.
type IPSCodeableConcept struct {
	Coding []IPSCoding `json:"coding,omitempty"`
	Text   string      `json:"text,omitempty"`
}

// IPSCoding representa un codigo en FHIR.
type IPSCoding struct {
	System  string `json:"system,omitempty"`
	Code    string `json:"code,omitempty"`
	Display string `json:"display,omitempty"`
}

// IPSReference representa una referencia en FHIR.
type IPSReference struct {
	Reference string `json:"reference"`
}

// IPSReaction representa una reaccion alergica en FHIR.
type IPSReaction struct {
	Manifestation []IPSCodeableConcept `json:"manifestation"`
	Severity      string               `json:"severity,omitempty"`
}

// ProjectAllergyToIPS proyecta un registro de alergia del Core
// al perfil IPS AllergyIntolerance (FHIR R4 simplificado).
//
// El Core almacena el contenido como blob opaco; este Shader lo interpreta.
// La proyeccion es responsabilidad del Shader, nunca del Core.
func ProjectAllergyToIPS(data ExportData) (*IPSAllergyIntolerance, error) {
	var content AllergyContent
	if err := json.Unmarshal([]byte(data.Content), &content); err != nil {
		return nil, fmt.Errorf("error al parsear contenido de alergia: %w", err)
	}
	if content.Type != "allergy" {
		return nil, fmt.Errorf("tipo de contenido incorrecto: %s", content.Type)
	}

	// Estado clinico: activo si issued, inactivo si voided
	clinicalStatusCode := "active"
	if data.State == "voided" {
		clinicalStatusCode = "inactive"
	}

	// Certeza → verificationStatus FHIR
	verificationCode := "confirmed"
	switch content.Certeza {
	case "sospecha":
		verificationCode = "unconfirmed"
	case "descartada":
		verificationCode = "refuted"
	}

	// Criticidad → criticality FHIR
	criticality := ""
	switch content.Criticidad {
	case "leve":
		criticality = "low"
	case "moderada":
		criticality = "low"
	case "grave":
		criticality = "high"
	}

	// Fecha de registro
	recordedDate := data.CreatedAt.Format(time.RFC3339)
	if data.IssuedAt != nil {
		recordedDate = data.IssuedAt.Format(time.RFC3339)
	}

	resource := &IPSAllergyIntolerance{
		ResourceType: "AllergyIntolerance",
		ID:           data.EvidenceID,
		ClinicalStatus: IPSCodeableConcept{
			Coding: []IPSCoding{{
				System:  "http://terminology.hl7.org/CodeSystem/allergyintolerance-clinical",
				Code:    clinicalStatusCode,
				Display: clinicalStatusCode,
			}},
		},
		VerificationStatus: IPSCodeableConcept{
			Coding: []IPSCoding{{
				System:  "http://terminology.hl7.org/CodeSystem/allergyintolerance-verification",
				Code:    verificationCode,
				Display: verificationCode,
			}},
		},
		Type:        "allergy",
		Criticality: criticality,
		Code: IPSCodeableConcept{
			Text: content.Agente,
		},
		Patient: IPSReference{
			Reference: "Patient/" + data.SubjectRef,
		},
		RecordedDate: recordedDate,
	}

	// Reaccion
	if content.TipoReaccion != "" {
		severity := ""
		switch content.Criticidad {
		case "leve":
			severity = "mild"
		case "moderada":
			severity = "moderate"
		case "grave":
			severity = "severe"
		}
		resource.Reaction = []IPSReaction{{
			Manifestation: []IPSCodeableConcept{{Text: content.TipoReaccion}},
			Severity:      severity,
		}}
	}

	return resource, nil
}

// ExportAllergyAsIPS serializa una alergia como JSON IPS.
// El resultado es efimero — no se persiste (ADR-0007).
func ExportAllergyAsIPS(data ExportData) ([]byte, error) {
	resource, err := ProjectAllergyToIPS(data)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(resource, "", "  ")
}
