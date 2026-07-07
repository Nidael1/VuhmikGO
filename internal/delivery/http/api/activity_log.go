package api

import (
	"context"
	"time"

	"github.com/Nidael1/VuhmikGO/internal/observability"
)

// logActivity registra un evento en activity_log (ADR-0023).
// Sin PHI — solo tenant_id, tipo de evento y timestamp.
// Fallo no bloquea el flujo principal.
func logActivity(ctx context.Context, tenantID, eventType string) {
	if deps.DB == nil || tenantID == "" {
		return
	}
	id := eventType + "-" + time.Now().UTC().Format("20060102150405.000000000")
	_, err := deps.DB.Exec(ctx, `
		INSERT INTO activity_log (id, tenant_id, event_type, occurred_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO NOTHING
	`, id, tenantID, eventType, time.Now().UTC())
	if err != nil {
		observability.Logger.Error("activity_log: error al registrar evento",
			"event_type", eventType,
			"error", err.Error(),
		)
	}
}
