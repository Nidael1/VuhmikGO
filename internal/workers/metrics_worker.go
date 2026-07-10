package workers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Nidael1/VuhmikGO/internal/observability"
)

// MetricsWorker calcula el snapshot de metricas de negocio (ADR-0019).
// WAR-A: corre cada 4 horas. Escribe en metrics_snapshot.
// Solo lee datos agregados — nunca PHI ni contenido clinico.
type MetricsWorker struct {
	interval time.Duration
	pool     *pgxpool.Pool
}

// NewMetricsWorker crea el worker de calculo de metricas.
func NewMetricsWorker(pool *pgxpool.Pool) *MetricsWorker {
	return &MetricsWorker{
		interval: 4 * time.Hour,
		pool:     pool,
	}
}

// Start arranca el worker en background. Bloquea hasta que el contexto se cancele.
func (w *MetricsWorker) Start(ctx context.Context) {
	observability.Logger.Info("metrics worker iniciado", "interval", w.interval)

	// Calculo inmediato al arrancar
	if err := w.run(ctx); err != nil {
		observability.Logger.Error("error en calculo inicial de metricas", "error", err.Error())
	}

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			observability.Logger.Info("metrics worker detenido")
			return
		case <-ticker.C:
			if err := w.run(ctx); err != nil {
				observability.Logger.Error("error en calculo periodico de metricas", "error", err.Error())
			}
		}
	}
}

// accountDetail es el detalle por cuenta para el snapshot.
type accountDetail struct {
	TenantID    string  `json:"tenant_id"`
	Email       string  `json:"email"`
	State       string  `json:"state"`
	MRR         float64 `json:"mrr"`
	Patients    int     `json:"patients"`
	LastRecord  string  `json:"last_record,omitempty"`
}

// run ejecuta el calculo completo y escribe un nuevo snapshot.
// Todos los datos son agregados — sin PHI ni contenido clinico.
// Calculate ejecuta el calculo de metricas inmediatamente (recalculo manual desde panel admin).
func (w *MetricsWorker) Calculate() error {
	return w.run(context.Background())
}

func (w *MetricsWorker) run(ctx context.Context) error {
	// 1. Conteos globales de cuentas
	var totalAccounts, activeAccounts, suspendedAccounts int
	err := w.pool.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE is_suspended = false),
			COUNT(*) FILTER (WHERE is_suspended = true)
		FROM users
		WHERE is_admin = false
	`).Scan(&totalAccounts, &activeAccounts, &suspendedAccounts)
	if err != nil {
		return err
	}

	// 2. MRR: suma de cuotas mensuales fijas + suma de costos por módulo activo
	// según el billing_mode de cada tenant (ADR-0025, issue #238)
	var mrr float64
	err = w.pool.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(CASE WHEN u.billing_mode = 'monthly' THEN u.monthly_fee ELSE 0 END), 0) +
			COALESCE(SUM(CASE WHEN u.billing_mode = 'per_module' OR u.billing_mode IS NULL
				THEN (SELECT COALESCE(SUM(tc2.costo), 0)
				      FROM tenant_capabilities tc2
				      WHERE tc2.tenant_id = u.tenant_id AND tc2.active = true)
				ELSE 0 END), 0)
		FROM users u
		WHERE u.is_admin = false
	`).Scan(&mrr)
	if err != nil {
		return err
	}

	// 3. Conteos globales de registros
	var totalPatients, totalNotes, totalAllergies, totalPrescriptions int
	err = w.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM patients),
			(SELECT COUNT(*) FROM note_projections  WHERE state = 'issued'),
			(SELECT COUNT(*) FROM allergy_projections WHERE state = 'issued'),
			(SELECT COUNT(*) FROM prescription_projections WHERE state = 'issued')
	`).Scan(&totalPatients, &totalNotes, &totalAllergies, &totalPrescriptions)
	if err != nil {
		return err
	}

	// 4. Detalle por cuenta (sin PHI: solo tenant_id, email, estado, MRR, conteo pacientes)
	rows, err := w.pool.Query(ctx, `
		SELECT
			u.tenant_id,
			u.email,
			CASE WHEN u.is_suspended THEN 'suspended' ELSE 'active' END AS state,
			CASE WHEN u.billing_mode = 'monthly'
				THEN u.monthly_fee
				ELSE COALESCE((SELECT SUM(tc2.costo) FROM tenant_capabilities tc2 WHERE tc2.tenant_id = u.tenant_id AND tc2.active = true), 0)
			END AS mrr,
			COUNT(DISTINCT p.id) AS patients,
			MAX(p.created_at)::text AS last_record
		FROM users u
		LEFT JOIN patients p ON p.tenant_id = u.tenant_id
		WHERE u.is_admin = false
		GROUP BY u.tenant_id, u.email, u.is_suspended, u.billing_mode, u.monthly_fee
		ORDER BY patients DESC
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	accounts := make([]accountDetail, 0)
	for rows.Next() {
		var a accountDetail
		if err := rows.Scan(&a.TenantID, &a.Email, &a.State, &a.MRR, &a.Patients, &a.LastRecord); err != nil {
			continue
		}
		accounts = append(accounts, a)
	}

	// 5. Distribucion de modulos
	modRows, err := w.pool.Query(ctx, `
		SELECT module_id, COUNT(*) AS total
		FROM tenant_capabilities
		WHERE active = true
		GROUP BY module_id
		ORDER BY total DESC
	`)
	if err != nil {
		return err
	}
	defer modRows.Close()

	modulesDistribution := make(map[string]int)
	for modRows.Next() {
		var key string
		var count int
		if err := modRows.Scan(&key, &count); err != nil {
			continue
		}
		modulesDistribution[key] = count
	}

	// 6. Serializar y escribir snapshot
	accountsJSON, err := json.Marshal(accounts)
	if err != nil {
		return err
	}
	modulesJSON, err := json.Marshal(modulesDistribution)
	if err != nil {
		return err
	}

	_, err = w.pool.Exec(ctx, `
		INSERT INTO metrics_snapshot (
			calculated_at, total_accounts, active_accounts, suspended_accounts,
			mrr, total_patients, total_notes, total_allergies, total_prescriptions,
			accounts_detail, modules_distribution
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`,
		time.Now().UTC(),
		totalAccounts, activeAccounts, suspendedAccounts,
		mrr,
		totalPatients, totalNotes, totalAllergies, totalPrescriptions,
		string(accountsJSON),
		string(modulesJSON),
	)
	if err != nil {
		return err
	}

	observability.Logger.Info("metrics snapshot calculado",
		"total_accounts", totalAccounts,
		"active_accounts", activeAccounts,
		"mrr", mrr,
		"total_patients", totalPatients,
	)
	return nil
}
