package shaders

import (
	"encoding/json"
	"fmt"
	"time"
)

// IPSCondition representa el recurso Condition del IPS (ADR-0010, ADR-0013).
// Usado para Problem List (activos) e History of Past Illness (resueltos).
// Perfil FHIR R4 simplificado.
type IPSCondition struct {
	ResourceType      string             `json:"resourceType"`
	ID                string             `json:"id"`
	ClinicalStatus    IPSCodeableConcept `json:"clinicalStatus"`
	VerificationStatus IPSCodeableConcept `json:"verificationStatus"`
	Category          []IPSCodeableConcept `json:"category"`
	Code              IPSCodeableConcept `json:"code"`
	Subject           IPSReference       `json:"subject"`
	OnsetDateTime     string             `json:"onsetDateTime,omitempty"`
	RecordedDate      string             `json:"recordedDate"`
}

// ProjectDiagnosisToIPS proyecta un registro de diagnóstico del Core
// al recurso FHIR Condition (ADR-0010, ADR-0013).
func ProjectDiagnosisToIPS(data ExportData) (*IPSCondition, error) {
	var content DiagnosisContent
	if err := json.Unmarshal([]byte(data.Content), &content); err != nil {
		return nil, fmt.Errorf("error al parsear contenido de diagnóstico: %w", err)
	}
	if content.Type != "diagnosis" {
		return nil, fmt.Errorf("tipo de contenido incorrecto: %s", content.Type)
	}

	// Estado clínico según estado del problema
	clinicalCode := "active"
	clinicalDisplay := "Active"
	if content.EstadoProblema == "resuelto" {
		clinicalCode = "resolved"
		clinicalDisplay = "Resolved"
	} else if content.EstadoProblema == "cronico" {
		clinicalCode = "active"
		clinicalDisplay = "Active"
	}

	// Categoría IPS: problem-list-item
	categoryCode := "problem-list-item"
	if content.EstadoProblema == "resuelto" {
		categoryCode = "encounter-diagnosis"
	}

	// Código CIE-10 si existe, texto libre si no
	codeableConcept := IPSCodeableConcept{Text: content.Descripcion}
	if content.CodigoCIE10 != "" {
		codeableConcept.Coding = []IPSCoding{{
			System:  "http://hl7.org/fhir/sid/icd-10",
			Code:    content.CodigoCIE10,
			Display: content.Descripcion,
		}}
	}

	recordedDate := data.CreatedAt.Format(time.RFC3339)
	if data.IssuedAt != nil {
		recordedDate = data.IssuedAt.Format(time.RFC3339)
	}

	condition := &IPSCondition{
		ResourceType: "Condition",
		ID:           data.EvidenceID,
		ClinicalStatus: IPSCodeableConcept{
			Coding: []IPSCoding{{
				System:  "http://terminology.hl7.org/CodeSystem/condition-clinical",
				Code:    clinicalCode,
				Display: clinicalDisplay,
			}},
		},
		VerificationStatus: IPSCodeableConcept{
			Coding: []IPSCoding{{
				System:  "http://terminology.hl7.org/CodeSystem/condition-ver-status",
				Code:    "confirmed",
				Display: "Confirmed",
			}},
		},
		Category: []IPSCodeableConcept{{
			Coding: []IPSCoding{{
				System:  "http://terminology.hl7.org/CodeSystem/condition-category",
				Code:    categoryCode,
				Display: categoryCode,
			}},
		}},
		Code:         codeableConcept,
		Subject:      IPSReference{Reference: "Patient/" + data.SubjectRef},
		OnsetDateTime: content.FechaInicio,
		RecordedDate: recordedDate,
	}

	return condition, nil
}

// ExportDiagnosisAsIPS serializa un diagnóstico como JSON IPS Condition.
func ExportDiagnosisAsIPS(data ExportData) ([]byte, error) {
	condition, err := ProjectDiagnosisToIPS(data)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(condition, "", "  ")
}
