package observability

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

const (
	maxFileSizeBytes = 10 * 1024 * 1024
	maxBackupFiles   = 5
)

type FileLogger struct {
	mu           sync.Mutex
	activityFile *os.File
	errorsFile   *os.File
	logDir       string
	activitySize int64
	errorsSize   int64
}

func NewFileLogger(logDir string) (*FileLogger, error) {
	if logDir == "" { return nil, nil }
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("error al crear directorio de logs: %w", err)
	}
	fl := &FileLogger{logDir: logDir}
	return fl, fl.openFiles()
}

func (fl *FileLogger) openFiles() error {
	af, err := os.OpenFile(filepath.Join(fl.logDir, "vuhmik-activity.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil { return fmt.Errorf("error al abrir activity log: %w", err) }
	ef, err := os.OpenFile(filepath.Join(fl.logDir, "vuhmik-errors.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil { af.Close(); return fmt.Errorf("error al abrir errors log: %w", err) }
	if info, err := af.Stat(); err == nil { fl.activitySize = info.Size() }
	if info, err := ef.Stat(); err == nil { fl.errorsSize = info.Size() }
	fl.activityFile = af
	fl.errorsFile = ef
	return nil
}

func (fl *FileLogger) write(f **os.File, size *int64, name, line string) {
	if *f == nil { return }
	if *size >= maxFileSizeBytes {
		(*f).Close()
		fl.rotate(name)
		nf, err := os.OpenFile(filepath.Join(fl.logDir, name), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil { return }
		*f = nf; *size = 0
	}
	data := []byte(line + "\n")
	n, err := (*f).Write(data)
	if err == nil { *size += int64(n) }
}

func (fl *FileLogger) rotate(name string) {
	os.Remove(filepath.Join(fl.logDir, fmt.Sprintf("%s.%d", name, maxBackupFiles)))
	for i := maxBackupFiles - 1; i >= 1; i-- {
		os.Rename(
			filepath.Join(fl.logDir, fmt.Sprintf("%s.%d", name, i)),
			filepath.Join(fl.logDir, fmt.Sprintf("%s.%d", name, i+1)),
		)
	}
	os.Rename(filepath.Join(fl.logDir, name), filepath.Join(fl.logDir, name+".1"))
}

func (fl *FileLogger) WriteActivity(line string) { fl.mu.Lock(); defer fl.mu.Unlock(); fl.write(&fl.activityFile, &fl.activitySize, "vuhmik-activity.log", line) }
func (fl *FileLogger) WriteError(line string)    { fl.mu.Lock(); defer fl.mu.Unlock(); fl.write(&fl.errorsFile, &fl.errorsSize, "vuhmik-errors.log", line) }
func (fl *FileLogger) Close() {
	fl.mu.Lock(); defer fl.mu.Unlock()
	if fl.activityFile != nil { fl.activityFile.Close() }
	if fl.errorsFile != nil { fl.errorsFile.Close() }
}

type dualHandler struct {
	stdout slog.Handler
	fl     *FileLogger
}

func (h *dualHandler) Enabled(ctx context.Context, level slog.Level) bool { return h.stdout.Enabled(ctx, level) }

func (h *dualHandler) Handle(ctx context.Context, r slog.Record) error {
	if err := h.stdout.Handle(ctx, r); err != nil { return err }
	line := fmt.Sprintf("%s [%s] %s", r.Time.Format("2006-01-02 15:04:05"), r.Level.String(), r.Message)
	r.Attrs(func(a slog.Attr) bool { line += fmt.Sprintf(" %s=%v", a.Key, a.Value); return true })
	h.fl.WriteActivity(line)
	if r.Level >= slog.LevelWarn { h.fl.WriteError(line) }
	return nil
}

func (h *dualHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return &dualHandler{stdout: h.stdout.WithAttrs(attrs), fl: h.fl} }
func (h *dualHandler) WithGroup(name string) slog.Handler       { return &dualHandler{stdout: h.stdout.WithGroup(name), fl: h.fl} }

func InitFileLogging() func() {
	logDir := os.Getenv("LOG_DIR")
	fl, err := NewFileLogger(logDir)
	if err != nil { slog.Error("error al inicializar file logger", "error", err); return func() {} }
	if fl == nil { return func() {} }
	slog.SetDefault(slog.New(&dualHandler{stdout: slog.Default().Handler(), fl: fl}))
	slog.Info("file logging iniciado", "log_dir", logDir)
	return fl.Close
}
