// Package postgres provee adaptadores PostgreSQL para los repositorios.
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// User representa un usuario del sistema en PostgreSQL.
type User struct {
	CURP         string
	ID           string
	TenantID     string
	Email        string
	PasswordHash string
	IsAdmin      bool
	IsSuspended  bool
	CreatedAt    time.Time
}

// UserRepository es el adaptador PostgreSQL para usuarios.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository retorna un repositorio de usuarios PostgreSQL.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// Create inserta un nuevo usuario en la BD.
func (r *UserRepository) Create(u User) error {
	sql := `
		INSERT INTO users (id, tenant_id, email, password_hash, curp, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.pool.Exec(context.Background(), sql,
		u.ID, u.TenantID, u.Email, u.PasswordHash, u.CURP, u.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("error al crear usuario: %w", err)
	}
	return nil
}

// FindByEmail busca un usuario por email.
// Retorna error si no existe.
func (r *UserRepository) FindByEmail(email string) (User, error) {
	sql := `
		SELECT id, tenant_id, email, password_hash, curp, is_admin, is_suspended, created_at
		FROM users WHERE email = $1`
	row := r.pool.QueryRow(context.Background(), sql, email)
	var u User
	if err := row.Scan(&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.CURP, &u.IsAdmin, &u.IsSuspended, &u.CreatedAt); err != nil {
		return User{}, fmt.Errorf("usuario no encontrado: %w", err)
	}
	return u, nil
}

// ExistsByEmail verifica si un email ya está registrado.
func (r *UserRepository) ExistsByEmail(email string) bool {
	var count int
	sql := `SELECT COUNT(1) FROM users WHERE email = $1`
	r.pool.QueryRow(context.Background(), sql, email).Scan(&count)
	return count > 0
}

// FindByID busca un usuario por su ID.
func (r *UserRepository) FindByID(id string) (User, error) {
	sql := `
		SELECT id, tenant_id, email, password_hash, curp, is_admin, is_suspended, created_at
		FROM users WHERE id = $1`
	row := r.pool.QueryRow(context.Background(), sql, id)
	var u User
	if err := row.Scan(&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.CURP, &u.IsAdmin, &u.IsSuspended, &u.CreatedAt); err != nil {
		return User{}, fmt.Errorf("usuario no encontrado: %w", err)
	}
	return u, nil
}
