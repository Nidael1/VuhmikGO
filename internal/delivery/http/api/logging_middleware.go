package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/Nidael1/VuhmikGO/internal/auth"
)

// responseWriter captura el status code para loguearlo.
type responseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytes += n
	return n, err
}

// LoggingMiddleware registra cada request HTTP con metodo, path, status y duracion.
// Sin PHI/PII — no loguea body ni headers de autorizacion.
// tenant_id extraido del JWT para trazabilidad sin datos clinicos.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		duration := time.Since(start)

		attrs := []any{
			"method",   r.Method,
			"path",     r.URL.Path,
			"status",   rw.status,
			"ms",       duration.Milliseconds(),
			"bytes",    rw.bytes,
		}

		if claims, ok := r.Context().Value(claimsKey{}).(*auth.Claims); ok && claims != nil {
			attrs = append(attrs, "tenant_id", claims.TenantID)
		}

		switch {
		case rw.status >= 500:
			slog.Error("request", attrs...)
		case rw.status >= 400:
			slog.Warn("request", attrs...)
		default:
			slog.Info("request", attrs...)
		}
	})
}
