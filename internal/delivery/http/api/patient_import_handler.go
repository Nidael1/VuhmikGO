package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Nidael1/VuhmikGO/internal/core/evidence"
	"github.com/Nidael1/VuhmikGO/internal/infrastructure/postgres"
)

// TransferPackage es el formato del paquete de traspaso (ADR-0009).
// El médico destino importa este JSON generado por el médico origen.
type TransferPackage struct {
	Format      string                   `json:"format"`
	Version     string                   `json:"version"`
	Hash        string                   `json:"hash"`
	ExportedAt  string                   `json:"exported_at"`
	Patient     TransferPatient          `json:"patient"`
	Evidence    []TransferEvidenceRecord `json:"evidence"`
}

// TransferPatient datos del paciente en el paquete de traspaso.
type TransferPatient struct {
	Nombre          string `json:"nombre"`
	FechaNacimiento string `json:"fecha_nacimiento"`
	Sexo            string `json:"sexo"`
	CURP            string `json:"curp"`
}

// TransferEvidenceRecord un registro de evidencia en el paquete de traspaso.
type TransferEvidenceRecord struct {
	OriginalID string `json:"original_id"`
	Content    string `json:"content"`
	IssuedAt   string `json:"issued_at"`
	Type       string `json:"type"`
}

// ImportResult resultado de la importación del paquete.
type ImportResult struct {
	PatientID       string   `json:"patient_id"`
	Created         bool     `json:"created"`
	DuplicateCURP   bool     `json:"duplicate_curp"`
	EvidenceImported int     `json:"evidence_imported"`
	Warnings        []string `json:"warnings,omitempty"`
}

// HandlePatientImport importa un expediente de traspaso (ADR-0009).
// Verifica el hash de integridad, busca por CURP, crea o alerta de duplicado.
//
// POST /api/v1/patients/import
func HandlePatientImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := TenantIDFromContext(r)
	actorID := ActorIDFromContext(r)
	if tenantID == "" || actorID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}

	var pkg TransferPackage
	if err := json.NewDecoder(r.Body).Decode(&pkg); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "payload invalido")
		return
	}

	// Validar formato del paquete
	if pkg.Format != "vuhmik-transfer-v1" {
		writeError(w, http.StatusBadRequest, "INVALID_FORMAT", "formato de traspaso no reconocido")
		return
	}

	// Verificar hash de integridad (ADR-0008, ADR-0009)
	if pkg.Hash != "" {
		if err := verifyTransferHash(pkg); err != nil {
			writeError(w, http.StatusUnprocessableEntity, "HASH_MISMATCH",
				"el paquete de traspaso no pasa la verificación de integridad")
			return
		}
	}

	warnings := []string{}

	// Buscar si el CURP ya existe en el tenant destino
	var patientID string
	duplicateCURP := false
	created := false

	existingPatient, err := deps.PatientRepo.FindByCURP(tenantID, pkg.Patient.CURP)
	if err == nil && existingPatient.ID != "" {
		// CURP ya existe en el tenant destino — no crear duplicado
		patientID = existingPatient.ID
		duplicateCURP = true
		warnings = append(warnings, fmt.Sprintf(
			"paciente con CURP %s ya existe (ID: %s) — evidencia importada sobre registro existente",
			pkg.Patient.CURP, existingPatient.ID,
		))
	} else {
		// CURP no existe — crear nuevo paciente
		now := time.Now().UTC()
		numExp, _ := deps.PatientRepo.NextExpediente(tenantID)
		patientID = "pat-import-" + now.Format("20060102150405.000")

		curp := strings.ToUpper(strings.TrimSpace(pkg.Patient.CURP))
		newPatient := postgres.Patient{
			ID:              patientID,
			TenantID:        tenantID,
			Nombre:          pkg.Patient.Nombre,
			FechaNacimiento: pkg.Patient.FechaNacimiento,
			Sexo:            pkg.Patient.Sexo,
			NumExpediente:   numExp,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if curp != "" {
			newPatient.CURP = &curp
		}
		if err := deps.PatientRepo.Create(newPatient); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al crear paciente importado")
			return
		}
		created = true
	}

	// Importar evidencias como registros issued con referencia al origen (ADR-0009)
	imported := 0
	for _, rec := range pkg.Evidence {
		if rec.Content == "" {
			continue
		}

		// Enriquecer el blob con metadatos de import
		var blob map[string]any
		if err := json.Unmarshal([]byte(rec.Content), &blob); err != nil {
			warnings = append(warnings, fmt.Sprintf("registro %s: blob invalido, omitido", rec.OriginalID))
			continue
		}
		blob["import_source"] = "vuhmik-transfer-v1"
		blob["import_ref"] = rec.OriginalID
		enrichedContent, err := json.Marshal(blob)
		if err != nil {
			continue
		}

		now := time.Now().UTC()
		issuedAt := now
		if rec.IssuedAt != "" {
			if t, err := time.Parse(time.RFC3339, rec.IssuedAt); err == nil {
				issuedAt = t
			}
		}
		id := "ev-import-" + now.Format("20060102150405.000") + "-" + rec.OriginalID[:min(8, len(rec.OriginalID))]

		e := evidence.Evidence{
			ID:         id,
			TenantID:   tenantID,
			SubjectRef: patientID,
			Content:    string(enrichedContent),
			State:      evidence.StateIssued, // Los registros importados llegan emitidos (ADR-0009)
			CreatedAt:  now,
			IssuedAt:   &issuedAt,
		}
		if err := deps.EvidenceRepo.Create(e); err != nil {
			warnings = append(warnings, fmt.Sprintf("registro %s: error al importar", rec.OriginalID))
			continue
		}
		imported++
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"data": ImportResult{
			PatientID:        patientID,
			Created:          created,
			DuplicateCURP:    duplicateCURP,
			EvidenceImported: imported,
			Warnings:         warnings,
		},
		"error": nil,
	})
}

// verifyTransferHash verifica la integridad del paquete de traspaso (ADR-0008).
func verifyTransferHash(pkg TransferPackage) error {
	saved := pkg.Hash
	pkg.Hash = ""
	b, err := json.Marshal(pkg)
	if err != nil {
		return err
	}
	h := sha256.Sum256(b)
	computed := "sha256:" + hex.EncodeToString(h[:])
	if computed != saved {
		return fmt.Errorf("hash mismatch: esperado %s, calculado %s", saved, computed)
	}
	return nil
}

// min retorna el mínimo de dos enteros.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
