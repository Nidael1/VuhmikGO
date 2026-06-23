package api

import (
	"github.com/Nidael1/VuhmikGO/internal/application/ports"
	"github.com/Nidael1/VuhmikGO/internal/infrastructure/postgres"
	infraredis "github.com/Nidael1/VuhmikGO/internal/infrastructure/redis"
)

// Deps contiene las dependencias inyectadas en los handlers de la API.
type Deps struct {
	EvidenceRepo     ports.EvidenceRepository
	UserRepo         *postgres.UserRepository
	PatientRepo      *postgres.PatientRepository
	RefreshTokenRepo *postgres.RefreshTokenRepository
	RedisClient      *infraredis.Client
}

var deps Deps

// InitDeps inicializa las dependencias de la API.
func InitDeps(d Deps) {
	deps = d
}
