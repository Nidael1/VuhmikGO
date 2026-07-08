package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"

	delivery "github.com/Nidael1/VuhmikGO/internal/delivery/http"
	"github.com/Nidael1/VuhmikGO/internal/application"
	"github.com/Nidael1/VuhmikGO/internal/delivery/http/api"
	"github.com/Nidael1/VuhmikGO/internal/infrastructure/postgres"
	infraredis "github.com/Nidael1/VuhmikGO/internal/infrastructure/redis"
	"github.com/Nidael1/VuhmikGO/internal/observability"
	"github.com/Nidael1/VuhmikGO/internal/workers"
)

func main() {
	if err := observability.ValidateRuntimeSecrets(); err != nil {
		log.Fatalf("error de configuración: %v", err)
	}

	// PostgreSQL
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("error al conectar PostgreSQL: %v", err)
	}
	defer pool.Close()
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("error al verificar conexion PostgreSQL: %v", err)
	}

	// Redis (WAR-A obligatorio)
	redisClient, err := infraredis.NewClient()
	if err != nil {
		log.Fatalf("error al conectar Redis: %v", err)
	}
	defer redisClient.Close()
	observability.Logger.Info("redis conectado")

	// Inyectar dependencias
	capabilityRepo := postgres.NewCapabilityRepository(pool)
	api.InitDeps(api.Deps{
		DB: pool,
		EvidenceRepo:               postgres.NewEvidenceRepository(pool),
		UserRepo:                   postgres.NewUserRepository(pool),
		PatientRepo:                postgres.NewPatientRepository(pool),
		RefreshTokenRepo:           postgres.NewRefreshTokenRepository(pool),
		RedisClient:                redisClient,
		CapabilityRepo:             capabilityRepo,
		AllergyService:             application.NewAllergyService(postgres.NewEvidenceRepository(pool), postgres.NewAllergyProjectionRepository(pool), capabilityRepo),
		AllergyProjectionRepo:      postgres.NewAllergyProjectionRepository(pool),
		NoteProjectionRepo:         postgres.NewNoteProjectionRepository(pool),
		PrescriptionProjectionRepo: postgres.NewPrescriptionProjectionRepository(pool),
		ConsultationProjectionRepo: postgres.NewConsultationProjectionRepository(pool),
		ConsultationService:        application.NewConsultationService(postgres.NewEvidenceRepository(pool), postgres.NewConsultationProjectionRepository(pool)),
		PrescriptionService:        application.NewPrescriptionService(postgres.NewEvidenceRepository(pool), postgres.NewPrescriptionProjectionRepository(pool), capabilityRepo),
		DiagnosisService:           application.NewDiagnosisService(postgres.NewEvidenceRepository(pool), postgres.NewDiagnosisProjectionRepository(pool), capabilityRepo),
		DiagnosisProjectionRepo:    postgres.NewDiagnosisProjectionRepository(pool),
		ImmunizationService:        application.NewImmunizationService(postgres.NewEvidenceRepository(pool), postgres.NewImmunizationProjectionRepository(pool), capabilityRepo),
		ImmunizationProjectionRepo: postgres.NewImmunizationProjectionRepository(pool),
		LabResultService:           application.NewLabResultService(postgres.NewEvidenceRepository(pool), postgres.NewLabResultProjectionRepository(pool), capabilityRepo),
		LabResultProjectionRepo:    postgres.NewLabResultProjectionRepository(pool),
		ProfileRepo:                postgres.NewProfileRepository(pool),
		TenantRepo:                 postgres.NewTenantRepository(pool),
		VendorRepo:                 postgres.NewVendorRepository(pool),
	})

	mux := http.NewServeMux()
	delivery.InitDeliveryDeps(delivery.DeliveryDeps{
		TenantRepo: postgres.NewTenantRepository(pool),
	})
	delivery.RegisterRoutes(mux)
	api.RegisterAPIRoutes(mux)
	handler := delivery.Handler(mux)

	// Contexto global para workers WAR-A
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Workers WAR-A obligatorios
	go workers.NewBackupWorker().Start(ctx)
	go workers.NewMetricsPurgeWorker().Start(ctx)
	go workers.NewMetricsWorker(pool).Start(ctx)

	// Señal de shutdown graceful
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	observability.Logger.Info("servidor iniciado", "addr", ":8080")

	go func() {
		if err := http.ListenAndServe(":8080", handler); err != nil && err != http.ErrServerClosed {
			log.Fatalf("error al iniciar servidor: %v", err)
		}
	}()

	<-quit
	observability.Logger.Info("servidor detenido")
	cancel()
}
