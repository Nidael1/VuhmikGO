package observability

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	maxFileSizeBytes = 10 * 1024 * 1024
	maxBackupFiles   = 5
)

// FileLogger escribe logs a dos archivos:
// - vuhmik-activity-YYYY-MM-DD.log (uno por dia, todo el trafico)
// - vuhmik-errors.log (acumula, solo WARN y ERROR, rota por tamano)
type FileLogger struct {
	mu           sync.Mutex
	activityFile *os.File
	errorsFile   *os.File
	logDir       string
	activityDay  string
	errorsSize   int64
}

func NewFileLogger(logDir string) (*FileLogger, error) {
	if logDir == "" {
		return nil, nil
	}
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("error al crear directorio de logs: %w", err)
	}
	fl := &FileLogger{logDir: logDir}
	return fl, fl.openFiles()
}

func (fl *FileLogger) activityFileName(day string) string {
	return filepath.Join(fl.logDir, fmt.Sprintf("vuhmik-activity-%s.log", day))
}

func (fl *FileLogger) openFiles() error {
	today := time.Now().Format("2006-01-02")
	af, err := os.OpenFile(fl.activityFileName(today), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error al abrir activity log: %w", err)
	}
	ef, err := os.OpenFile(filepath.Join(fl.logDir, "vuhmik-errors.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		af.Close()
		return fmt.Errorf("error al abrir errors log: %w", err)
	}
	if info, err := ef.Stat(); err == nil {
		fl.errorsSize = info.Size()
	}
	fl.activityFile = af
	fl.activityDay = today
	fl.errorsFile = ef
	return nil
}

// WriteActivity escribe en el archivo del dia actual.
// Si cambia la fecha, cierra el archivo anterior y abre uno nuevo.
func (fl *FileLogger) WriteActivity(line string) {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	today := time.Now().Format("2006-01-02")
	if today != fl.activityDay {
		if fl.activityFile != nil {
			fl.activityFile.Close()
		}
		nf, err := os.OpenFile(fl.activityFileName(today), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return
		}
		fl.activityFile = nf
		fl.activityDay = today
	}
	fl.activityFile.Write([]byte(line + "\n"))
}

// WriteError escribe en vuhmik-errors.log con rotacion por tamano (10MB).
func (fl *FileLogger) WriteError(line string) {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	if fl.errorsFile == nil {
		return
	}
	if fl.errorsSize >= maxFileSizeBytes {
		fl.errorsFile.Close()
		fl.rotateErrors()
		nf, err := os.OpenFile(filepath.Join(fl.logDir, "vuhmik-errors.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return
		}
		fl.errorsFile = nf
		fl.errorsSize = 0
	}
	data := []byte(line + "\n")
	n, err := fl.errorsFile.Write(data)
	if err == nil {
		fl.errorsSize += int64(n)
	}
}

func (fl *FileLogger) rotateErrors() {
	name := "vuhmik-errors.log"
	os.Remove(filepath.Join(fl.logDir, fmt.Sprintf("%s.%d", name, maxBackupFiles)))
	for i := maxBackupFiles - 1; i >= 1; i-- {
		os.Rename(
			filepath.Join(fl.logDir, fmt.Sprintf("%s.%d", name, i)),
			filepath.Join(fl.logDir, fmt.Sprintf("%s.%d", name, i+1)),
		)
	}
	os.Rename(filepath.Join(fl.logDir, name), filepath.Join(fl.logDir, name+".1"))
}

func (fl *FileLogger) Close() {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	if fl.activityFile != nil {
		fl.activityFile.Close()
	}
	if fl.errorsFile != nil {
		fl.errorsFile.Close()
	}
}

type dualHandler struct {
	stdout slog.Handler
	fl     *FileLogger
}

func (h *dualHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.stdout.Enabled(ctx, level)
}

func (h *dualHandler) Handle(ctx context.Context, r slog.Record) error {
	if err := h.stdout.Handle(ctx, r); err != nil {
		return err
	}
	line := fmt.Sprintf("%s [%s] %s", r.Time.Format("2006-01-02 15:04:05"), r.Level.String(), r.Message)
	r.Attrs(func(a slog.Attr) bool {
		line += fmt.Sprintf(" %s=%v", a.Key, a.Value)
		return true
	})
	h.fl.WriteActivity(line)
	if r.Level >= slog.LevelWarn {
		h.fl.WriteError(line)
	}
	return nil
}

func (h *dualHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &dualHandler{stdout: h.stdout.WithAttrs(attrs), fl: h.fl}
}

func (h *dualHandler) WithGroup(name string) slog.Handler {
	return &dualHandler{stdout: h.stdout.WithGroup(name), fl: h.fl}
}

// InitFileLogging configura slog para escribir a stdout y archivos si LOG_DIR esta definido.
func InitFileLogging() func() {
	logDir := os.Getenv("LOG_DIR")
	fl, err := NewFileLogger(logDir)
	if err != nil {
		slog.Error("error al inicializar file logger", "error", err)
		return func() {}
	}
	if fl == nil {
		return func() {}
	}
	slog.SetDefault(slog.New(&dualHandler{
		stdout: slog.Default().Handler(),
		fl:     fl,
	}))
	slog.Info("file logging iniciado", "log_dir", logDir)
	return fl.Close
}
