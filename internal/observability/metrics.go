package observability

import "sync/atomic"

// EventMetrics mantiene contadores agregados por tipo de evento.
//
// Reglas absolutas:
//   - Sin IDs de documentos, pacientes ni tenants.
//   - Sin datos clínicos ni personales.
//   - Solo contadores anónimos por categoría de operación.
//   - Thread-safe mediante sync/atomic.
type EventMetrics struct {
	Create  atomic.Int64
	Issue   atomic.Int64
	Void    atomic.Int64
	Replace atomic.Int64
	Export  atomic.Int64
	Errors  atomic.Int64
}

// GlobalMetrics es el registro central de métricas del sistema.
// Un solo punto de conteo — sin duplicados ni contadores locales.
var GlobalMetrics = &EventMetrics{}

// RecordOperation incrementa el contador de la operación dada.
// Si failed es true, incrementa el contador de errores.
// No acepta IDs ni datos personales.
func RecordOperation(op string, failed bool) {
	if failed {
		GlobalMetrics.Errors.Add(1)
		return
	}
	switch op {
	case "create":
		GlobalMetrics.Create.Add(1)
	case "issue":
		GlobalMetrics.Issue.Add(1)
	case "void":
		GlobalMetrics.Void.Add(1)
	case "replace":
		GlobalMetrics.Replace.Add(1)
	case "export":
		GlobalMetrics.Export.Add(1)
	}
}

// Snapshot retorna una copia inmutable de los contadores actuales.
// Seguro para lectura concurrente. No modifica el estado global.
func Snapshot() map[string]int64 {
	return map[string]int64{
		"create":  GlobalMetrics.Create.Load(),
		"issue":   GlobalMetrics.Issue.Load(),
		"void":    GlobalMetrics.Void.Load(),
		"replace": GlobalMetrics.Replace.Load(),
		"export":  GlobalMetrics.Export.Load(),
		"errors":  GlobalMetrics.Errors.Load(),
	}
}

// ResetMetrics resetea todos los contadores de metricas en memoria.
// Llamado por el MetricsPurgeWorker cada 30 dias (WAR-A).
func ResetMetrics() {
	GlobalMetrics.Create.Store(0)
	GlobalMetrics.Issue.Store(0)
	GlobalMetrics.Void.Store(0)
	GlobalMetrics.Replace.Store(0)
	GlobalMetrics.Export.Store(0)
	GlobalMetrics.Errors.Store(0)
}
