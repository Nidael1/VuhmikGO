package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	delivery "github.com/Nidael1/VuhmikGO/internal/delivery/http"
	"github.com/Nidael1/VuhmikGO/internal/application"
	"github.com/Nidael1/VuhmikGO/internal/delivery/http/api"
	"github.com/Nidael1/VuhmikGO/internal/infrastructure/postgres"
	infraredis "github.com/Nidael1/VuhmikGO/internal/infrastructure/redis"
	"github.com/Nidael1/VuhmikGO/internal/observability"
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
		EvidenceRepo:     postgres.NewEvidenceRepository(pool),
		UserRepo:         postgres.NewUserRepository(pool),
		PatientRepo:      postgres.NewPatientRepository(pool),
		RefreshTokenRepo: postgres.NewRefreshTokenRepository(pool),
		RedisClient:      redisClient,
		CapabilityRepo:   capabilityRepo,
		AllergyService:        application.NewAllergyService(postgres.NewEvidenceRepository(pool), postgres.NewAllergyProjectionRepository(pool), capabilityRepo),
		AllergyProjectionRepo: postgres.NewAllergyProjectionRepository(pool),
		NoteProjectionRepo:    postgres.NewNoteProjectionRepository(pool),
		ProfileRepo:      postgres.NewProfileRepository(pool),
	})

	mux := http.NewServeMux()
	delivery.RegisterRoutes(mux)
	api.RegisterAPIRoutes(mux)
	handler := delivery.Handler(mux)

	observability.Logger.Info("servidor iniciado", "addr", ":8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("error al iniciar servidor: %v", err)
	}
}
