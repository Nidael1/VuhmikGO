package workers

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SystemWorker struct {
	pool *pgxpool.Pool
}

func NewSystemWorker(pool *pgxpool.Pool) *SystemWorker {
	return &SystemWorker{pool: pool}
}

func (w *SystemWorker) Start(ctx context.Context) {
	slog.Info("system worker iniciado", "interval", time.Hour.String())
	if err := w.Calculate(); err != nil {
		slog.Error("error en calculo inicial de sistema", "error", err)
	}
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			slog.Info("system worker detenido")
			return
		case <-ticker.C:
			if err := w.Calculate(); err != nil {
				slog.Error("error en calculo de sistema", "error", err)
			}
		}
	}
}

func (w *SystemWorker) Calculate() error {
	ctx := context.Background()
	if err := w.calculateAccountHealth(ctx); err != nil {
		return fmt.Errorf("account health: %w", err)
	}
	if err := w.calculateSystemSnapshot(ctx); err != nil {
		return fmt.Errorf("system snapshot: %w", err)
	}
	slog.Info("system snapshot calculado")
	return nil
}

func (w *SystemWorker) calculateAccountHealth(ctx context.Context) error {
	now := time.Now().UTC()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	firstOfLastMonth := firstOfMonth.AddDate(0, -1, 0)

	rows, err := w.pool.Query(ctx, `
		SELECT
			u.tenant_id, u.email, u.created_at,
			(SELECT MAX(occurred_at) FROM activity_log
			 WHERE tenant_id = u.tenant_id AND event_type = 'session_start') AS last_login_at,
			(SELECT COUNT(*) FROM activity_log
			 WHERE tenant_id = u.tenant_id AND event_type = 'session_start'
			 AND occurred_at >= $1) AS sessions_this_month,
			(SELECT COUNT(*) FROM activity_log
			 WHERE tenant_id = u.tenant_id AND event_type = 'session_start'
			 AND occurred_at >= $2 AND occurred_at < $1) AS sessions_last_month,
			(SELECT COUNT(*) FROM note_projections
			 WHERE tenant_id = u.tenant_id AND state = 'issued' AND issued_at >= $1) AS notes_this_month,
			(SELECT COUNT(*) FROM prescription_projections
			 WHERE tenant_id = u.tenant_id AND state = 'issued' AND issued_at >= $1) AS prescriptions_this_month,
			(SELECT COUNT(*) FROM allergy_projections
			 WHERE tenant_id = u.tenant_id AND state = 'issued' AND issued_at >= $1) AS allergies_this_month,
			(SELECT COUNT(*) FROM note_projections
			 WHERE tenant_id = u.tenant_id AND state = 'issued') AS total_notes,
			(SELECT COUNT(*) FROM prescription_projections
			 WHERE tenant_id = u.tenant_id AND state = 'issued') AS total_prescriptions,
			(SELECT COUNT(*) FROM patients WHERE tenant_id = u.tenant_id) AS total_patients,
			(SELECT COUNT(*) FROM tenant_capabilities
			 WHERE tenant_id = u.tenant_id AND active = true) AS modules_active
		FROM users u WHERE u.is_admin = false ORDER BY u.created_at ASC
	`, firstOfMonth, firstOfLastMonth)
	if err != nil {
		return fmt.Errorf("error al consultar salud: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			tenantID, email              string
			createdAt                    time.Time
			lastLoginAt                  *time.Time
			sessThisMonth, sessLastMonth  int
			notesM, prescM, allergM      int
			totalNotes, totalPresc, totalPat int
			modulesActive                int
		)
		if err := rows.Scan(&tenantID, &email, &createdAt, &lastLoginAt,
			&sessThisMonth, &sessLastMonth, &notesM, &prescM, &allergM,
			&totalNotes, &totalPresc, &totalPat, &modulesActive); err != nil {
			return err
		}
		daysSinceLogin := 999
		if lastLoginAt != nil {
			daysSinceLogin = int(now.Sub(*lastLoginAt).Hours() / 24)
		}
		ageDays := int(now.Sub(createdAt).Hours() / 24)
		modulesUsed := 0
		if notesM > 0 { modulesUsed++ }
		if prescM > 0 { modulesUsed++ }
		if allergM > 0 { modulesUsed++ }
		healthStatus := "active"
		if daysSinceLogin > 30 {
			healthStatus = "inactive"
		} else if daysSinceLogin > 7 || (notesM == 0 && ageDays > 14) {
			healthStatus = "at_risk"
		}
		_, err = w.pool.Exec(ctx, `
			INSERT INTO account_health_snapshot (
				tenant_id, email, calculated_at,
				account_age_days, last_login_at, days_since_login,
				sessions_this_month, sessions_last_month,
				notes_this_month, prescriptions_this_month, allergies_this_month,
				total_notes, total_prescriptions, total_patients,
				modules_active, modules_used, health_status
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
			ON CONFLICT (tenant_id) DO UPDATE SET
				email=$2, calculated_at=$3, account_age_days=$4,
				last_login_at=$5, days_since_login=$6,
				sessions_this_month=$7, sessions_last_month=$8,
				notes_this_month=$9, prescriptions_this_month=$10,
				allergies_this_month=$11, total_notes=$12,
				total_prescriptions=$13, total_patients=$14,
				modules_active=$15, modules_used=$16, health_status=$17
		`, tenantID, email, now, ageDays, lastLoginAt, daysSinceLogin,
			sessThisMonth, sessLastMonth, notesM, prescM, allergM,
			totalNotes, totalPresc, totalPat, modulesActive, modulesUsed, healthStatus)
		if err != nil {
			return fmt.Errorf("error al guardar salud %s: %w", tenantID, err)
		}
	}
	return nil
}

func (w *SystemWorker) calculateSystemSnapshot(ctx context.Context) error {
	now := time.Now().UTC()
	issues := ""
	overallOk := true

	dbOk := true
	if err := w.pool.QueryRow(ctx, `SELECT 1`).Scan(new(int)); err != nil {
		dbOk = false; overallOk = false; issues += "BD no responde. "
	}

	var lastBackupAt *time.Time
	lastBackupSizeKB := 0
	backupOk := false
	entries, err := os.ReadDir("/tmp/vuhmik-backups")
	if err == nil {
		var latestTime time.Time
		var latestSize int64
		for _, e := range entries {
			if e.IsDir() { continue }
			info, _ := e.Info()
			if info != nil && info.ModTime().After(latestTime) {
				latestTime = info.ModTime()
				latestSize = info.Size()
			}
		}
		if !latestTime.IsZero() {
			lastBackupAt = &latestTime
			lastBackupSizeKB = int(latestSize / 1024)
			backupOk = now.Sub(latestTime).Hours() < 25
			if !backupOk { overallOk = false; issues += "Backup desactualizado. " }
		} else {
			issues += "Sin backups encontrados. "
		}
	}

	var metricsLastRunAt *time.Time
	metricsOk := false
	var lastMetrics time.Time
	if err2 := w.pool.QueryRow(ctx, `SELECT calculated_at FROM metrics_snapshot ORDER BY calculated_at DESC LIMIT 1`).Scan(&lastMetrics); err2 == nil {
		metricsLastRunAt = &lastMetrics
		metricsOk = now.Sub(lastMetrics).Hours() < 5
		if !metricsOk { overallOk = false; issues += "Worker de métricas sin correr hace más de 5h. " }
	}

	diskUsedPct := 0
	diskOk := true
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/", &stat); err == nil {
		total := stat.Blocks * uint64(stat.Bsize)
		free := stat.Bfree * uint64(stat.Bsize)
		if total > 0 { diskUsedPct = int((total-free) * 100 / total) }
		diskOk = diskUsedPct < 85
		if !diskOk { overallOk = false; issues += fmt.Sprintf("Disco al %d%%. ", diskUsedPct) }
	}

	if issues == "" { issues = "Todo funciona correctamente." }
	_, _ = w.pool.Exec(ctx, `DELETE FROM system_snapshot WHERE id NOT IN (SELECT id FROM system_snapshot ORDER BY calculated_at DESC LIMIT 48)`)
	_, err3 := w.pool.Exec(ctx, `
		INSERT INTO system_snapshot (calculated_at, db_ok, last_backup_at, last_backup_size_kb, backup_ok,
			metrics_last_run_at, metrics_ok, disk_used_pct, disk_ok, overall_ok, issues)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`, now, dbOk, lastBackupAt, lastBackupSizeKB, backupOk, metricsLastRunAt, metricsOk, diskUsedPct, diskOk, overallOk, issues)
	return err3
}
