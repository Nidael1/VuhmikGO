package shaders

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"time"
)

// IPSBundle representa un IPS Document Bundle FHIR R4 mínimo.
// El Core nunca conoce esta estructura; vive exclusivamente en Shaders/export.
// Referencia: HL7 FHIR R4 / IPS Implementation Guide.
// ADR-0010: IPS sobre FHIR como modelo canónico de intercambio.
type IPSBundle struct {
	ResourceType string        `json:"resourceType" xml:"resourceType"`
	ID           string        `json:"id" xml:"id"`
	Meta         IPSMeta       `json:"meta" xml:"meta"`
	Type         string        `json:"type" xml:"type"`
	Timestamp    string        `json:"timestamp" xml:"timestamp"`
	Entry        []IPSEntry    `json:"entry" xml:"entry"`
}

// IPSMeta contiene metadatos del Bundle IPS.
type IPSMeta struct {
	Profile []string `json:"profile" xml:"profile"`
}

// IPSEntry representa una entrada del Bundle IPS.
type IPSEntry struct {
	FullURL  string      `json:"fullUrl" xml:"fullUrl"`
	Resource interface{} `json:"resource" xml:"resource"`
}

// IPSComposition es el recurso Composition del IPS.
// Es siempre la primera entrada del Bundle.
type IPSComposition struct {
	ResourceType string             `json:"resourceType"`
	ID           string             `json:"id"`
	Status       string             `json:"status"`
	Type         IPSCodeableConcept `json:"type"`
	Subject      IPSReference       `json:"subject"`
	Date         string             `json:"date"`
	Author       []IPSReference     `json:"author"`
	Title        string             `json:"title"`
	Section      []IPSSection       `json:"section"`
}

// IPSSection es una sección de la Composition IPS.
type IPSSection struct {
	Title string             `json:"title"`
	Code  IPSCodeableConcept `json:"code"`
	Text  IPSNarrative       `json:"text"`
	Entry []IPSReference     `json:"entry,omitempty"`
}

// IPSNarrative es el texto narrativo de una sección.
type IPSNarrative struct {
	Status string `json:"status"`
	Div    string `json:"div"`
}

// IPSPatient es un recurso Patient mínimo para el Bundle IPS.
type IPSPatient struct {
	ResourceType string `json:"resourceType"`
	ID           string `json:"id"`
}

// IPSNoteContent parsea el blob de una nota clínica genérica.
// El Core nunca lo interpreta; solo el Shader lo conoce.
type IPSNoteContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// BuildIPSBundle construye un IPS Bundle FHIR R4 mínimo a partir de ExportData.
// Soporta notas clínicas genéricas (type: "note") en v1.
// Las secciones de alergias, recetas y diagnósticos se agregan en issues posteriores
// conforme se implementan sus proyectores IPS.
// ADR-0010, issue #212.
func BuildIPSBundle(data ExportData, contentHash string) (*IPSBundle, error) {
	issuedAt := time.Now().UTC().Format(time.RFC3339)
	if data.IssuedAt != nil {
		issuedAt = data.IssuedAt.Format(time.RFC3339)
	}

	// Parsear blob de nota si el tipo es "note"
	noteText := data.Content
	var noteContent IPSNoteContent
	if err := json.Unmarshal([]byte(data.Content), &noteContent); err == nil {
		if noteContent.Text != "" {
			noteText = noteContent.Text
		}
	}

	patientRef := "Patient/" + data.SubjectRef
	compositionID := "composition-" + data.EvidenceID
	patientID := "patient-" + data.SubjectRef

	// Sección principal: nota clínica como narrativa IPS
	section := IPSSection{
		Title: "Historia Clínica",
		Code: IPSCodeableConcept{
			Coding: []IPSCoding{{
				System:  "http://loinc.org",
				Code:    "11329-0",
				Display: "History general",
			}},
		},
		Text: IPSNarrative{
			Status: "generated",
			Div:    fmt.Sprintf(`<div xmlns="http://www.w3.org/1999/xhtml">%s</div>`, noteText),
		},
	}

	composition := IPSComposition{
		ResourceType: "Composition",
		ID:           compositionID,
		Status:       "final",
		Type: IPSCodeableConcept{
			Coding: []IPSCoding{{
				System:  "http://loinc.org",
				Code:    "60591-5",
				Display: "Patient summary Document",
			}},
		},
		Subject: IPSReference{Reference: patientRef},
		Date:    issuedAt,
		Author:  []IPSReference{{Reference: "Organization/" + data.TenantID}},
		Title:   "Resumen de Paciente — VUHMÍK",
		Section: []IPSSection{section},
	}

	patient := IPSPatient{
		ResourceType: "Patient",
		ID:           patientID,
	}

	bundle := &IPSBundle{
		ResourceType: "Bundle",
		ID:           "ips-" + data.EvidenceID,
		Meta: IPSMeta{
			Profile: []string{
				"http://hl7.org/fhir/uv/ips/StructureDefinition/Bundle-uv-ips",
			},
		},
		Type:      "document",
		Timestamp: issuedAt,
		Entry: []IPSEntry{
			{FullURL: "urn:uuid:" + compositionID, Resource: composition},
			{FullURL: "urn:uuid:" + patientID, Resource: patient},
		},
	}

	return bundle, nil
}

// MarshalIPSBundleJSON serializa el Bundle IPS como JSON FHIR R4.
// Incluye el hash de integridad como extensión del Bundle (ADR-0008).
func MarshalIPSBundleJSON(bundle *IPSBundle, contentHash string) ([]byte, error) {
	// Representación con hash como mapa para incluirlo limpiamente
	type BundleWithHash struct {
		ResourceType string                 `json:"resourceType"`
		ID           string                 `json:"id"`
		Meta         IPSMeta                `json:"meta"`
		Type         string                 `json:"type"`
		Timestamp    string                 `json:"timestamp"`
		VuhmikHash   string                 `json:"_vuhmik_hash,omitempty"`
		Entry        []IPSEntry             `json:"entry"`
	}

	out := BundleWithHash{
		ResourceType: bundle.ResourceType,
		ID:           bundle.ID,
		Meta:         bundle.Meta,
		Type:         bundle.Type,
		Timestamp:    bundle.Timestamp,
		VuhmikHash:   contentHash,
		Entry:        bundle.Entry,
	}
	return json.MarshalIndent(out, "", "  ")
}

// MarshalIPSBundleXML serializa el Bundle IPS como XML FHIR R4.
// Reemplaza el esquema CDA propio (ADR-0007 legado → ADR-0010 canónico).
// ADR-0010: el esquema CDA de ADR-0007 queda deprecado; IPS/FHIR es el canónico.
func MarshalIPSBundleXML(bundle *IPSBundle, contentHash string) ([]byte, error) {
	type XMLEntry struct {
		FullURL  string `xml:"fullUrl"`
		Resource string `xml:"resource"`
	}
	type XMLBundle struct {
		XMLName      xml.Name `xml:"Bundle"`
		Xmlns        string   `xml:"xmlns,attr"`
		ID           string   `xml:"id"`
		Profile      string   `xml:"meta>profile"`
		Type         string   `xml:"type"`
		Timestamp    string   `xml:"timestamp"`
		VuhmikHash   string   `xml:"_vuhmik_hash,omitempty"`
		Composition  string   `xml:"entry>resource>Composition>title"`
		Section      string   `xml:"entry>resource>Composition>section>text>div"`
		Subject      string   `xml:"entry>resource>Composition>subject>reference"`
	}

	// Para XML FHIR real se requeriría un serializador complejo.
	// En v1 generamos XML FHIR-compatible simplificado con los campos
	// obligatorios del IPS Bundle, pendiente de expansión en issue posterior.
	type SimpleIPSXML struct {
		XMLName   xml.Name `xml:"Bundle"`
		Xmlns     string   `xml:"xmlns,attr"`
		ID        string   `xml:"id"`
		Profile   string   `xml:"meta>profile"`
		Type      string   `xml:"type"`
		Timestamp string   `xml:"timestamp"`
		Hash      string   `xml:"_vuhmik_hash"`
		Entries   []struct {
			FullURL xml.CharData `xml:"fullUrl"`
		} `xml:"entry"`
	}

	// Extraer narrativa de la composition para XML
	narrativeDiv := ""
	compositionTitle := "Resumen de Paciente — VUHMÍK"
	subjectRef := ""
	if len(bundle.Entry) > 0 {
		if comp, ok := bundle.Entry[0].Resource.(IPSComposition); ok {
			compositionTitle = comp.Title
			subjectRef = comp.Subject.Reference
			if len(comp.Section) > 0 {
				narrativeDiv = comp.Section[0].Text.Div
			}
		}
	}

	type IPSXMLFull struct {
		XMLName   xml.Name `xml:"Bundle"`
		Xmlns     string   `xml:"xmlns,attr"`
		ID        string   `xml:"id"`
		Profile   string   `xml:"meta>profile"`
		Type      string   `xml:"type"`
		Timestamp string   `xml:"timestamp"`
		Hash      string   `xml:"_vuhmik_hash"`
		Title     string   `xml:"entry>resource>Composition>title"`
		Subject   string   `xml:"entry>resource>Composition>subject>reference"`
		Section   string   `xml:"entry>resource>Composition>section>text>div"`
	}

	doc := IPSXMLFull{
		Xmlns:     "http://hl7.org/fhir",
		ID:        bundle.ID,
		Profile:   "http://hl7.org/fhir/uv/ips/StructureDefinition/Bundle-uv-ips",
		Type:      bundle.Type,
		Timestamp: bundle.Timestamp,
		Hash:      contentHash,
		Title:     compositionTitle,
		Subject:   subjectRef,
		Section:   narrativeDiv,
	}

	_ = narrativeDiv // usado arriba

	output, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error al generar XML IPS: %w", err)
	}
	header := []byte(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	return append(header, output...), nil
}
