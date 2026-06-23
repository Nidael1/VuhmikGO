package workers

import (
	"context"
	"time"

	"github.com/Nidael1/VuhmikGO/internal/observability"
)

// MetricsPurgeWorker purga metricas agregadas con mas de 30 dias.
// WAR-A obliga a retención fija — no series crudas ilimitadas.
// Este worker corre cada 24 horas y elimina registros antiguos.
type MetricsPurgeWorker struct {
	interval  time.Duration
	retention time.Duration
}

// NewMetricsPurgeWorker crea un worker de purge de metricas.
func NewMetricsPurgeWorker() *MetricsPurgeWorker {
	return &MetricsPurgeWorker{
		interval:  24 * time.Hour,
		retention: 30 * 24 * time.Hour, // 30 dias
	}
}

// Start arranca el worker en background.
func (w *MetricsPurgeWorker) Start(ctx context.Context) {
	observability.Logger.Info("metrics purge worker iniciado",
		"interval", w.interval,
		"retention", w.retention,
	)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			observability.Logger.Info("metrics purge worker detenido")
			return
		case <-ticker.C:
			w.run()
		}
	}
}

// run ejecuta el purge de metricas.
// Las metricas de VUHMIK v1 son contadores atomicos en memoria
// (internal/observability/metrics.go) — no persisten en BD.
// Este worker registra el evento de purge y resetea los contadores.
// En v2 con metricas en PostgreSQL, aqui se hara DELETE WHERE ts < cutoff.
func (w *MetricsPurgeWorker) run() {
	cutoff := time.Now().UTC().Add(-w.retention)
	observability.Logger.Info("metrics purge ejecutado",
		"cutoff", cutoff.Format(time.RFC3339),
		"note", "metricas en memoria reseteadas — v2 purgara BD",
	)
	// Resetear contadores en memoria
	observability.ResetMetrics()
}
