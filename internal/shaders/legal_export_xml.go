package shaders

import (
	"encoding/xml"
	"fmt"
	"time"
)

// ClinicalDocument es el documento XML de export legal.
// Basado en HL7 CDA simplificado adaptado a NOM-004-SSA3-2012.
// Namespace: urn:vuhmik:hl7:v1
type ClinicalDocument struct {
	XMLName    xml.Name        `xml:"ClinicalDocument"`
	Xmlns      string          `xml:"xmlns,attr"`
	ID         CDDocumentID    `xml:"id"`
	EffTime    CDEffectiveTime `xml:"effectiveTime"`
	Confidence CDConfidential  `xml:"confidentialityCode"`
	Record     CDRecordTarget  `xml:"recordTarget"`
	Author     CDAuthor        `xml:"author"`
	Component  CDComponent     `xml:"component"`
	Integrity  CDIntegrity     `xml:"integrity"`
}

type CDDocumentID struct {
	Root string `xml:"root,attr"`
}

type CDEffectiveTime struct {
	Value string `xml:"value,attr"`
}

type CDConfidential struct {
	Code string `xml:"code,attr"`
}

type CDRecordTarget struct {
	PatientRole CDPatientRole `xml:"patientRole"`
}

type CDPatientRole struct {
	ID CDDocumentID `xml:"id"`
}

type CDAuthor struct {
	Time         CDEffectiveTime `xml:"time"`
	AssignedAuth CDAssignedAuth  `xml:"assignedAuthor"`
}

type CDAssignedAuth struct {
	ID CDDocumentID `xml:"id"`
}

type CDComponent struct {
	StructuredBody CDStructuredBody `xml:"structuredBody"`
}

type CDStructuredBody struct {
	Component CDBodyComponent `xml:"component"`
}

type CDBodyComponent struct {
	Section CDSection `xml:"section"`
}

type CDSection struct {
	Title string `xml:"title"`
	Text  string `xml:"text"`
	State string `xml:"state"`
}

type CDIntegrity struct {
	Hash         CDHash  `xml:"hash"`
	State        string  `xml:"state"`
	ReplacedByID *string `xml:"replaced_by_id,omitempty"`
}

type CDHash struct {
	Algorithm string `xml:"algorithm,attr"`
	Value     string `xml:",chardata"`
}

// GenerateExportXML genera el export en formato XML (ADR-0007).
// El hash se calcula externamente y se pasa como parametro.
func GenerateExportXML(data ExportData, hash string) ([]byte, error) {
	issuedAt := ""
	if data.IssuedAt != nil {
		issuedAt = data.IssuedAt.Format(time.RFC3339)
	}

	doc := ClinicalDocument{
		Xmlns: "urn:vuhmik:hl7:v1",
		ID:    CDDocumentID{Root: data.EvidenceID},
		EffTime: CDEffectiveTime{
			Value: issuedAt,
		},
		Confidence: CDConfidential{Code: "N"},
		Record: CDRecordTarget{
			PatientRole: CDPatientRole{
				ID: CDDocumentID{Root: data.SubjectID},
			},
		},
		Author: CDAuthor{
			Time: CDEffectiveTime{
				Value: data.CreatedAt.Format(time.RFC3339),
			},
			AssignedAuth: CDAssignedAuth{
				ID: CDDocumentID{Root: data.TenantID},
			},
		},
		Component: CDComponent{
			StructuredBody: CDStructuredBody{
				Component: CDBodyComponent{
					Section: CDSection{
						Title: "Nota clinica",
						Text:  data.Notes,
						State: data.State,
					},
				},
			},
		},
		Integrity: CDIntegrity{
			Hash: CDHash{
				Algorithm: "SHA-256",
				Value:     hash,
			},
			State:        data.State,
			ReplacedByID: data.ReplacedByID,
		},
	}

	output, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error al generar XML: %w", err)
	}

	// Agregar cabecera XML
	header := []byte(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	return append(header, output...), nil
}
