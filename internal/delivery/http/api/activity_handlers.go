package api

import (
	"net/http"
	"strings"
)

// HandleAdminActivity lista todos los tenants con sus conteos de actividad.
// GET /api/v1/admin/activity
// Requiere AdminMiddleware. Solo lectura. Sin PHI.
func HandleAdminActivity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	if deps.DB == nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "db no disponible")
		return
	}

	rows, err := deps.DB.Query(r.Context(), `
		SELECT
			s.tenant_id,
			SUM(s.sessions_count)      AS sessions_total,
			SUM(s.notes_count)         AS notes_total,
			SUM(s.allergies_count)     AS allergies_total,
			SUM(s.prescriptions_count) AS prescriptions_total,
			SUM(s.exports_count)       AS exports_total,
			SUM(s.patients_count)      AS patients_total,
			MAX(s.period)::text        AS last_period
		FROM activity_snapshot s
		GROUP BY s.tenant_id
		ORDER BY sessions_total DESC
	`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al consultar actividad")
		return
	}
	defer rows.Close()

	type activityItem struct {
		TenantID          string `json:"tenant_id"`
		SessionsTotal     int    `json:"sessions_total"`
		NotesTotal        int    `json:"notes_total"`
		AllergiesTotal    int    `json:"allergies_total"`
		PrescriptionsTotal int   `json:"prescriptions_total"`
		ExportsTotal      int    `json:"exports_total"`
		PatientsTotal     int    `json:"patients_total"`
		LastPeriod        string `json:"last_period"`
	}

	items := make([]activityItem, 0)
	for rows.Next() {
		var item activityItem
		if err := rows.Scan(
			&item.TenantID,
			&item.SessionsTotal,
			&item.NotesTotal,
			&item.AllergiesTotal,
			&item.PrescriptionsTotal,
			&item.ExportsTotal,
			&item.PatientsTotal,
			&item.LastPeriod,
		); err != nil {
			continue
		}
		items = append(items, item)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data":  map[string]any{"items": items},
		"error": nil,
	})
}

// HandleAdminActivityDetail retorna el detalle de actividad de un tenant por mes.
// GET /api/v1/admin/activity/:tenant
// Requiere AdminMiddleware. Solo lectura. Sin PHI.
func HandleAdminActivityDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "metodo no permitido")
		return
	}
	tenantID := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/activity/")
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "tenant_id requerido")
		return
	}
	if deps.DB == nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "db no disponible")
		return
	}

	rows, err := deps.DB.Query(r.Context(), `
		SELECT
			period::text,
			sessions_count,
			notes_count,
			allergies_count,
			prescriptions_count,
			exports_count,
			patients_count
		FROM activity_snapshot
		WHERE tenant_id = $1
		ORDER BY period DESC
		LIMIT 12
	`, tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "error al consultar detalle")
		return
	}
	defer rows.Close()

	type periodItem struct {
		Period             string `json:"period"`
		SessionsCount      int    `json:"sessions_count"`
		NotesCount         int    `json:"notes_count"`
		AllergiesCount     int    `json:"allergies_count"`
		PrescriptionsCount int    `json:"prescriptions_count"`
		ExportsCount       int    `json:"exports_count"`
		PatientsCount      int    `json:"patients_count"`
	}

	periods := make([]periodItem, 0)
	for rows.Next() {
		var item periodItem
		if err := rows.Scan(
			&item.Period,
			&item.SessionsCount,
			&item.NotesCount,
			&item.AllergiesCount,
			&item.PrescriptionsCount,
			&item.ExportsCount,
			&item.PatientsCount,
		); err != nil {
			continue
		}
		periods = append(periods, item)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"tenant_id": tenantID,
			"periods":   periods,
		},
		"error": nil,
	})
}
