// Package workers contiene los workers de background de VUHMÍK (WAR-A).
// Los workers no manejan requests HTTP ni evidencia clínica directamente.
// Son procesos de mantenimiento, integridad y disponibilidad.
package workers

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/Nidael1/VuhmikGO/internal/observability"
)

// BackupWorker ejecuta backups automaticos de PostgreSQL cada 24 horas.
// Los backups se guardan en el directorio definido por BACKUP_DIR.
// En produccion, BACKUP_DIR debe apuntar a almacenamiento externo cifrado.
type BackupWorker struct {
	interval  time.Duration
	backupDir string
	dbURL     string
}

// NewBackupWorker crea un nuevo worker de backups.
func NewBackupWorker() *BackupWorker {
	backupDir := os.Getenv("BACKUP_DIR")
	if backupDir == "" {
		backupDir = "/tmp/vuhmik-backups"
	}
	return &BackupWorker{
		interval:  24 * time.Hour,
		backupDir: backupDir,
		dbURL:     os.Getenv("DATABASE_URL"),
	}
}

// Start arranca el worker en background. Bloquea hasta que el contexto se cancele.
func (w *BackupWorker) Start(ctx context.Context) {
	observability.Logger.Info("backup worker iniciado", "interval", w.interval)

	// Backup inmediato al arrancar
	if err := w.run(); err != nil {
		observability.Logger.Error("error en backup inicial", "error", err.Error())
	}

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			observability.Logger.Info("backup worker detenido")
			return
		case <-ticker.C:
			if err := w.run(); err != nil {
				observability.Logger.Error("error en backup periodico", "error", err.Error())
			}
		}
	}
}

// run ejecuta un pg_dump y guarda el archivo en BACKUP_DIR.
func (w *BackupWorker) run() error {
	if err := os.MkdirAll(w.backupDir, 0750); err != nil {
		return fmt.Errorf("error al crear directorio de backup: %w", err)
	}

	timestamp := time.Now().UTC().Format("20060102-150405")
	filename := filepath.Join(w.backupDir, fmt.Sprintf("vuhmik-backup-%s.sql", timestamp))

	cmd := exec.Command("pg_dump", "--no-password", "-f", filename, w.dbURL)
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error al ejecutar pg_dump: %w", err)
	}

	observability.Logger.Info("backup completado", "file", filename)

	// Purge de backups con mas de 7 dias
	w.purgeOldBackups()
	return nil
}

// purgeOldBackups elimina backups con mas de 7 dias de antiguedad.
func (w *BackupWorker) purgeOldBackups() {
	entries, err := os.ReadDir(w.backupDir)
	if err != nil {
		return
	}
	cutoff := time.Now().UTC().Add(-7 * 24 * time.Hour)
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			path := filepath.Join(w.backupDir, e.Name())
			os.Remove(path)
			observability.Logger.Info("backup purgado", "file", path)
		}
	}
}
