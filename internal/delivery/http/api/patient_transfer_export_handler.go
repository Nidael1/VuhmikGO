package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// HandlePatientExportTransfer genera el paquete de traspaso para un paciente.
// El formato TransferPackage es compatible con HandlePatientImport (ADR-0009).
// El archivo se genera en memoria y se sirve directamente. No se persiste.
// Cache-Control: no-store obligatorio.
//
// GET /api/v1/patients/:id/export/transfer
func HandlePatientExportTransfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := TenantIDFromContext(r)
	actorID := ActorIDFromContext(r)
	if tenantID == "" || actorID == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "no autenticado")
		return
	}

	patientID := strings.TrimPrefix(r.URL.Path, "/api/v1/patients/")
	patientID = strings.TrimSuffix(patientID, "/export/transfer")
	if patientID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "patient_id requerido")
		return
	}

	// Obtener datos del paciente
	p, err := deps.PatientRepo.FindByID(tenantID, patientID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "paciente no encontrado")
		return
	}

	// Obtener toda la evidencia activa del paciente (todos los tipos)
	allEvidence, err := deps.EvidenceRepo.FindAll(tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al obtener evidencia")
		return
	}

	records := make([]TransferEvidenceRecord, 0)
	for _, e := range allEvidence {
		if e.SubjectRef != patientID {
			continue
		}
		// Solo evidencia emitida — no drafts ni voided (ADR-0009)
		if e.State != "issued" {
			continue
		}
		rec := TransferEvidenceRecord{
			OriginalID: e.ID,
			Content:    e.Content,
		}
		if e.IssuedAt != nil {
			rec.IssuedAt = e.IssuedAt.Format(time.RFC3339)
		}
		// Extraer type del blob para el campo Type
		var blob map[string]any
		if err := json.Unmarshal([]byte(e.Content), &blob); err == nil {
			if t, ok := blob["type"].(string); ok {
				rec.Type = t
			}
		}
		records = append(records, rec)
	}

	// CURP del paciente
	curp := ""
	if p.CURP != nil {
		curp = *p.CURP
	}

	// Construir el TransferPackage sin hash todavía
	pkg := TransferPackage{
		Format:     "vuhmik-transfer-v1",
		Version:    "1.0",
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		Patient: TransferPatient{
			Nombre:          p.Nombre,
			FechaNacimiento: p.FechaNacimiento,
			Sexo:            p.Sexo,
			CURP:            curp,
		},
		Evidence: records,
	}

	// Calcular hash SHA-256 del paquete sin hash (ADR-0008, ADR-0009)
	pkgBytes, err := json.Marshal(pkg)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al serializar paquete")
		return
	}
	h := sha256.Sum256(pkgBytes)
	pkg.Hash = "sha256:" + hex.EncodeToString(h[:])

	// Serializar con hash incluido
	finalBytes, err := json.Marshal(pkg)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al serializar paquete final")
		return
	}

	filename := "traspaso_" + p.NumExpediente + "_" + time.Now().UTC().Format("20060102") + ".json"
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	w.Write(finalBytes)
}
