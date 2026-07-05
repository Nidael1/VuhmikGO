package shaders

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// AuditPackage representa el Audit Package ZIP por paciente.
// ADR-0027: ZIP síncrono en v1, generado en memoria, sin persistencia.
// Cache-Control: no-store obligatorio en el handler que lo sirva.
type AuditPackage struct {
	PatientID  string
	TenantID   string
	GeneratedAt time.Time
	Evidence   []ExportData
}

// manifestEntry es una entrada del manifest.json del ZIP.
type manifestEntry struct {
	File        string `json:"file"`
	Description string `json:"description"`
	Hash        string `json:"sha256"`
}

// auditManifest es el índice del paquete ZIP.
type auditManifest struct {
	Format      string          `json:"format"`
	Version     string          `json:"version"`
	PatientID   string          `json:"patient_id"`
	TenantID    string          `json:"tenant_id"`
	GeneratedAt string          `json:"generated_at"`
	Files       []manifestEntry `json:"files"`
	PackageHash string          `json:"package_hash"`
}

// BuildAuditPackageZIP construye el Audit Package ZIP en memoria.
// El resultado debe servirse directamente y descartarse.
// No se persiste en disco ni en base de datos (ADR-0027).
func BuildAuditPackageZIP(pkg AuditPackage) ([]byte, error) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	entries := []manifestEntry{}
	hashLines := []byte{}

	// 1) IPS Bundle FHIR R4 del primer registro de evidencia disponible
	//    (en v1 representa el resumen del paciente)
	if len(pkg.Evidence) > 0 {
		ipsData := pkg.Evidence[0]
		bundle, err := BuildIPSBundle(ipsData, "")
		if err == nil {
			ipsBytes, err := MarshalIPSBundleJSON(bundle, "")
			if err == nil {
				h := sha256.Sum256(ipsBytes)
				hashHex := hex.EncodeToString(h[:])
				filename := "ips/patient_ips.json"
				if err := writeZipFile(w, filename, ipsBytes); err != nil {
					return nil, fmt.Errorf("error al escribir IPS: %w", err)
				}
				entries = append(entries, manifestEntry{
					File:        filename,
					Description: "IPS Bundle FHIR R4 — resumen del paciente (ADR-0010)",
					Hash:        "sha256:" + hashHex,
				})
				hashLines = append(hashLines, []byte(hashHex+"  "+filename+"\n")...)
			}
		}
	}

	// 2) Cada registro de evidencia como JSON individual con hash
	for _, ev := range pkg.Evidence {
		h := sha256Hash(ev)
		evBytes, err := json.MarshalIndent(ev, "", "  ")
		if err != nil {
			continue
		}
		filename := fmt.Sprintf("evidence/record_%s.json", ev.EvidenceID)
		if err := writeZipFile(w, filename, evBytes); err != nil {
			return nil, fmt.Errorf("error al escribir evidencia %s: %w", ev.EvidenceID, err)
		}
		entries = append(entries, manifestEntry{
			File:        filename,
			Description: "Registro de evidencia — " + ev.EvidenceID,
			Hash:        "sha256:" + h,
		})
		hashLines = append(hashLines, []byte(h+"  "+filename+"\n")...)
	}

	// 3) hashes/hashes.sha256
	if err := writeZipFile(w, "hashes/hashes.sha256", hashLines); err != nil {
		return nil, fmt.Errorf("error al escribir hashes: %w", err)
	}

	// 4) manifest.json (sin hash propio — es el índice)
	manifest := auditManifest{
		Format:      "vuhmik-audit-package-v1",
		Version:     "1.0",
		PatientID:   pkg.PatientID,
		TenantID:    pkg.TenantID,
		GeneratedAt: pkg.GeneratedAt.Format(time.RFC3339),
		Files:       entries,
	}
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error al serializar manifest: %w", err)
	}
	if err := writeZipFile(w, "manifest.json", manifestBytes); err != nil {
		return nil, fmt.Errorf("error al escribir manifest: %w", err)
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("error al cerrar ZIP: %w", err)
	}

	return buf.Bytes(), nil
}

// writeZipFile agrega un archivo al ZIP writer.
func writeZipFile(w *zip.Writer, name string, data []byte) error {
	f, err := w.Create(name)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	return err
}

// sha256Hash calcula el hash SHA-256 de un ExportData serializado.
func sha256Hash(data ExportData) string {
	b, _ := json.Marshal(data)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
