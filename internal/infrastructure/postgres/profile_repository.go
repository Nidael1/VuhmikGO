package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/Nidael1/VuhmikGO/internal/application/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProfileRepository es el adaptador PostgreSQL para perfiles profesionales.
type ProfileRepository struct {
	pool *pgxpool.Pool
}

func NewProfileRepository(pool *pgxpool.Pool) *ProfileRepository {
	return &ProfileRepository{pool: pool}
}

func (r *ProfileRepository) Get(userID string) (ports.Profile, error) {
	sql := `
		SELECT user_id, tenant_id, rubro, nombre_completo, cedula_profesional, especialidad
		FROM professional_profiles
		WHERE user_id = $1 AND rubro = 'medicine'`
	var p ports.Profile
	err := r.pool.QueryRow(context.Background(), sql, userID).Scan(
		&p.UserID, &p.TenantID, &p.Rubro,
		&p.NombreCompleto, &p.CedulaProfesional, &p.Especialidad,
	)
	if err != nil {
		return ports.Profile{}, fmt.Errorf("perfil no encontrado: %w", err)
	}
	return p, nil
}

func (r *ProfileRepository) Upsert(p ports.Profile) error {
	sql := `
		INSERT INTO professional_profiles
			(user_id, tenant_id, rubro, nombre_completo, cedula_profesional, especialidad, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		ON CONFLICT (user_id, rubro) DO UPDATE SET
			nombre_completo    = EXCLUDED.nombre_completo,
			cedula_profesional = EXCLUDED.cedula_profesional,
			especialidad       = EXCLUDED.especialidad,
			updated_at         = EXCLUDED.updated_at`
	_, err := r.pool.Exec(context.Background(), sql,
		p.UserID, p.TenantID, p.Rubro,
		p.NombreCompleto, p.CedulaProfesional, p.Especialidad,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("error al guardar perfil: %w", err)
	}
	return nil
}
