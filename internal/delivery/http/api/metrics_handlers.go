package api

import (
	"net/http"
	"strings"
	"time"
)

// HandleAdminMetrics devuelve el ultimo snapshot de metricas agregadas.
// GET /api/v1/admin/metrics
// Requiere AdminMiddleware. Solo lectura. Sin PHI.
func HandleAdminMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	if deps.DB == nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "db no disponible")
		return
	}
	var snap struct {
		CalculatedAt       time.Time `json:"calculated_at"`
		TotalAccounts      int     `json:"total_accounts"`
		ActiveAccounts     int     `json:"active_accounts"`
		SuspendedAccounts  int     `json:"suspended_accounts"`
		MRR                float64 `json:"mrr"`
		TotalPatients      int     `json:"total_patients"`
		TotalNotes         int     `json:"total_notes"`
		TotalAllergies     int     `json:"total_allergies"`
		TotalPrescriptions int     `json:"total_prescriptions"`
	}
	err := deps.DB.QueryRow(r.Context(), `
		SELECT calculated_at, total_accounts, active_accounts, suspended_accounts,
		       mrr::float8, total_patients, total_notes, total_allergies, total_prescriptions
		FROM metrics_snapshot
		ORDER BY calculated_at DESC
		LIMIT 1
	`).Scan(
		&snap.CalculatedAt, &snap.TotalAccounts, &snap.ActiveAccounts,
		&snap.SuspendedAccounts, &snap.MRR, &snap.TotalPatients,
		&snap.TotalNotes, &snap.TotalAllergies, &snap.TotalPrescriptions,
	)
	if err != nil {
		writeError(w, http.StatusNotFound, "NO_SNAPSHOT", "no hay snapshot disponible aun")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": snap, "error": nil})
}

// HandleAdminMetricsAccounts devuelve la lista de cuentas con conteos desde el snapshot.
// GET /api/v1/admin/metrics/accounts
// Requiere AdminMiddleware. Solo lectura. Sin PHI.
func HandleAdminMetricsAccounts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	if deps.DB == nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "db no disponible")
		return
	}
	var calculatedAt time.Time
	var accountsJSON []byte
	err := deps.DB.QueryRow(r.Context(), `
		SELECT calculated_at, accounts_detail
		FROM metrics_snapshot
		ORDER BY calculated_at DESC
		LIMIT 1
	`).Scan(&calculatedAt, &accountsJSON)
	if err != nil {
		writeError(w, http.StatusNotFound, "NO_SNAPSHOT", "no hay snapshot disponible aun")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"calculated_at": calculatedAt,
			"accounts":      string(accountsJSON),
		},
		"error": nil,
	})
}

// HandleAdminMetricsAccountDetail devuelve el detalle de una cuenta desde el snapshot.
// GET /api/v1/admin/metrics/accounts/:id
// Requiere AdminMiddleware. Solo lectura. Sin PHI.
func HandleAdminMetricsAccountDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/metrics/accounts/")
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "tenant_id requerido")
		return
	}
	if deps.DB == nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "db no disponible")
		return
	}
	var calculatedAt time.Time
	var accountsJSON []byte
	err := deps.DB.QueryRow(r.Context(), `
		SELECT calculated_at, accounts_detail
		FROM metrics_snapshot
		ORDER BY calculated_at DESC
		LIMIT 1
	`).Scan(&calculatedAt, &accountsJSON)
	if err != nil {
		writeError(w, http.StatusNotFound, "NO_SNAPSHOT", "no hay snapshot disponible aun")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"calculated_at": calculatedAt,
			"tenant_id":     tenantID,
			"accounts":      string(accountsJSON),
		},
		"error": nil,
	})
}

// HandleAdminMetricsModules devuelve la distribucion de uso por modulo desde el snapshot.
// GET /api/v1/admin/metrics/modules
// Requiere AdminMiddleware. Solo lectura. Sin PHI.
func HandleAdminMetricsModules(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	if deps.DB == nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "db no disponible")
		return
	}
	var calculatedAt time.Time
	var modulesJSON []byte
	err := deps.DB.QueryRow(r.Context(), `
		SELECT calculated_at, modules_distribution
		FROM metrics_snapshot
		ORDER BY calculated_at DESC
		LIMIT 1
	`).Scan(&calculatedAt, &modulesJSON)
	if err != nil {
		writeError(w, http.StatusNotFound, "NO_SNAPSHOT", "no hay snapshot disponible aun")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"calculated_at": calculatedAt,
			"modules":       string(modulesJSON),
		},
		"error": nil,
	})
}
