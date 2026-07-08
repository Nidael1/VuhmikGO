package api

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Nidael1/VuhmikGO/internal/application"
	"github.com/Nidael1/VuhmikGO/internal/application/ports"
	"github.com/Nidael1/VuhmikGO/internal/infrastructure/postgres"
	infraredis "github.com/Nidael1/VuhmikGO/internal/infrastructure/redis"
)

// Deps contiene las dependencias inyectadas en los handlers de la API.
type Deps struct {
	DB               *pgxpool.Pool
	EvidenceRepo     ports.EvidenceRepository
	UserRepo         *postgres.UserRepository
	PatientRepo      *postgres.PatientRepository
	RefreshTokenRepo *postgres.RefreshTokenRepository
	RedisClient      *infraredis.Client
	CapabilityRepo   ports.CapabilityRepository
	AllergyService             *application.AllergyService
	AllergyProjectionRepo      ports.AllergyProjectionRepository
	ProfileRepo                ports.ProfileRepository
	NoteProjectionRepo         ports.NoteProjectionRepository
	NoteService                *application.NoteService
	PrescriptionService        *application.PrescriptionService
	PrescriptionProjectionRepo ports.PrescriptionProjectionRepository
	DiagnosisService           *application.DiagnosisService
	DiagnosisProjectionRepo    ports.DiagnosisProjectionRepository
	ImmunizationService        *application.ImmunizationService
	ImmunizationProjectionRepo ports.ImmunizationProjectionRepository
	LabResultService           *application.LabResultService
	LabResultProjectionRepo    ports.LabResultProjectionRepository
	ConsultationService        *application.ConsultationService
	ConsultationProjectionRepo ports.ConsultationProjectionRepository
	TenantRepo                 ports.TenantRepository
	VendorRepo                 ports.VendorRepository
}

var deps Deps

// InitDeps inicializa las dependencias de la API.
func InitDeps(d Deps) {
	deps = d
}
