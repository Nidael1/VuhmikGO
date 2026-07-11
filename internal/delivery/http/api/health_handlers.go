package api

import (
	"net/http"
	"strings"
	"time"
)

func HandleAdminAccountHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	statusFilter := r.URL.Query().Get("status")
	sql := `
		SELECT tenant_id, email, calculated_at,
			account_age_days, last_login_at, days_since_login,
			sessions_this_month, sessions_last_month,
			notes_this_month, prescriptions_this_month, allergies_this_month,
			total_notes, total_prescriptions, total_patients,
			modules_active, modules_used, health_status
		FROM account_health_snapshot`
	args := []any{}
	if statusFilter == "active" || statusFilter == "at_risk" || statusFilter == "inactive" {
		sql += ` WHERE health_status = $1`
		args = append(args, statusFilter)
	}
	sql += ` ORDER BY CASE health_status WHEN 'inactive' THEN 0 WHEN 'at_risk' THEN 1 ELSE 2 END, days_since_login DESC`

	rows, err := deps.DB.Query(r.Context(), sql, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al consultar salud")
		return
	}
	defer rows.Close()

	type AccountHealth struct {
		TenantID               string     `json:"tenant_id"`
		Email                  string     `json:"email"`
		CalculatedAt           time.Time  `json:"calculated_at"`
		AccountAgeDays         int        `json:"account_age_days"`
		LastLoginAt            *time.Time `json:"last_login_at"`
		DaysSinceLogin         int        `json:"days_since_login"`
		SessionsThisMonth      int        `json:"sessions_this_month"`
		SessionsLastMonth      int        `json:"sessions_last_month"`
		NotesThisMonth         int        `json:"notes_this_month"`
		PrescriptionsThisMonth int        `json:"prescriptions_this_month"`
		AllergiesThisMonth     int        `json:"allergies_this_month"`
		TotalNotes             int        `json:"total_notes"`
		TotalPrescriptions     int        `json:"total_prescriptions"`
		TotalPatients          int        `json:"total_patients"`
		ModulesActive          int        `json:"modules_active"`
		ModulesUsed            int        `json:"modules_used"`
		HealthStatus           string     `json:"health_status"`
	}

	var items []AccountHealth
	for rows.Next() {
		var a AccountHealth
		if err := rows.Scan(
			&a.TenantID, &a.Email, &a.CalculatedAt,
			&a.AccountAgeDays, &a.LastLoginAt, &a.DaysSinceLogin,
			&a.SessionsThisMonth, &a.SessionsLastMonth,
			&a.NotesThisMonth, &a.PrescriptionsThisMonth, &a.AllergiesThisMonth,
			&a.TotalNotes, &a.TotalPrescriptions, &a.TotalPatients,
			&a.ModulesActive, &a.ModulesUsed, &a.HealthStatus,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "SCAN_ERROR", err.Error())
			return
		}
		items = append(items, a)
	}
	if items == nil {
		items = []AccountHealth{}
	}
	active, atRisk, inactive := 0, 0, 0
	for _, a := range items {
		switch a.HealthStatus {
		case "active":   active++
		case "at_risk":  atRisk++
		case "inactive": inactive++
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"items":   items,
			"summary": map[string]int{"active": active, "at_risk": atRisk, "inactive": inactive},
		},
		"error": nil,
	})
}

func HandleAdminSystem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	var snap struct {
		CalculatedAt     time.Time  `json:"calculated_at"`
		DbOk             bool       `json:"db_ok"`
		LastBackupAt     *time.Time `json:"last_backup_at"`
		LastBackupSizeKB int        `json:"last_backup_size_kb"`
		BackupOk         bool       `json:"backup_ok"`
		MetricsLastRunAt *time.Time `json:"metrics_last_run_at"`
		MetricsOk        bool       `json:"metrics_ok"`
		DiskUsedPct      int        `json:"disk_used_pct"`
		DiskOk           bool       `json:"disk_ok"`
		OverallOk        bool       `json:"overall_ok"`
		Issues           string     `json:"issues"`
	}
	err := deps.DB.QueryRow(r.Context(), `
		SELECT calculated_at, db_ok, last_backup_at, last_backup_size_kb, backup_ok,
		       metrics_last_run_at, metrics_ok, disk_used_pct, disk_ok, overall_ok, issues
		FROM system_snapshot ORDER BY calculated_at DESC LIMIT 1
	`).Scan(
		&snap.CalculatedAt, &snap.DbOk,
		&snap.LastBackupAt, &snap.LastBackupSizeKB, &snap.BackupOk,
		&snap.MetricsLastRunAt, &snap.MetricsOk,
		&snap.DiskUsedPct, &snap.DiskOk,
		&snap.OverallOk, &snap.Issues,
	)
	if err != nil {
		writeError(w, http.StatusNotFound, "NO_SNAPSHOT", "sin datos de sistema aun")
		return
	}
	type LoginAttempt struct {
		Email      string    `json:"email"`
		OccurredAt time.Time `json:"occurred_at"`
	}
	failRows, _ := deps.DB.Query(r.Context(), `
		SELECT email, occurred_at FROM login_attempts ORDER BY occurred_at DESC LIMIT 20
	`)
	var failedLogins []LoginAttempt
	if failRows != nil {
		defer failRows.Close()
		for failRows.Next() {
			var f LoginAttempt
			if err := failRows.Scan(&f.Email, &f.OccurredAt); err == nil {
				failedLogins = append(failedLogins, f)
			}
		}
	}
	if failedLogins == nil {
		failedLogins = []LoginAttempt{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{"system": snap, "failed_logins": failedLogins},
		"error": nil,
	})
}

func HandleAdminSystemRecalculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	if deps.SystemWorker == nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "worker no disponible")
		return
	}
	if err := deps.SystemWorker.Calculate(); err != nil {
		writeError(w, http.StatusInternalServerError, "RECALCULATE_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"ok": true}, "error": nil})
}

func adminHealthDispatcher(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/health/")
	switch path {
	case "accounts":
		HandleAdminAccountHealth(w, r)
	default:
		writeError(w, http.StatusNotFound, "NOT_FOUND", "ruta no encontrada")
	}
}
