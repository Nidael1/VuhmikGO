package postgres

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RefreshToken representa un token de refresco stateful.
type RefreshToken struct {
	ID        string
	UserID    string
	TenantID  string
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

// RefreshTokenRepository es el adaptador PostgreSQL para refresh tokens.
type RefreshTokenRepository struct {
	pool *pgxpool.Pool
}

func NewRefreshTokenRepository(pool *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{pool: pool}
}

// GenerateToken genera un token aleatorio seguro y retorna
// el token en texto plano (para el cliente) y su hash (para la BD).
func GenerateRefreshTokenValue() (plain, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", fmt.Errorf("error al generar token: %w", err)
	}
	plain = hex.EncodeToString(b)
	h := sha256.Sum256([]byte(plain))
	hash = hex.EncodeToString(h[:])
	return plain, hash, nil
}

// Create persiste un nuevo refresh token en BD.
// Solo se guarda el hash — nunca el token en texto plano.
func (r *RefreshTokenRepository) Create(rt RefreshToken) error {
	sql := `
		INSERT INTO refresh_tokens (id, user_id, tenant_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.pool.Exec(context.Background(), sql,
		rt.ID, rt.UserID, rt.TenantID, rt.TokenHash, rt.ExpiresAt, rt.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("error al crear refresh token: %w", err)
	}
	return nil
}

// FindByHash busca un refresh token por su hash.
// Retorna error si no existe, está revocado o expirado.
func (r *RefreshTokenRepository) FindByHash(hash string) (RefreshToken, error) {
	sql := `
		SELECT id, user_id, tenant_id, token_hash, expires_at, revoked_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1`
	row := r.pool.QueryRow(context.Background(), sql, hash)
	var rt RefreshToken
	if err := row.Scan(
		&rt.ID, &rt.UserID, &rt.TenantID, &rt.TokenHash,
		&rt.ExpiresAt, &rt.RevokedAt, &rt.CreatedAt,
	); err != nil {
		return RefreshToken{}, fmt.Errorf("refresh token no encontrado: %w", err)
	}
	if rt.RevokedAt != nil {
		return RefreshToken{}, fmt.Errorf("refresh token revocado")
	}
	if time.Now().UTC().After(rt.ExpiresAt) {
		return RefreshToken{}, fmt.Errorf("refresh token expirado")
	}
	return rt, nil
}

// Revoke marca un refresh token como revocado.
func (r *RefreshTokenRepository) Revoke(id string) error {
	now := time.Now().UTC()
	sql := `UPDATE refresh_tokens SET revoked_at = $1 WHERE id = $2`
	_, err := r.pool.Exec(context.Background(), sql, now, id)
	if err != nil {
		return fmt.Errorf("error al revocar refresh token: %w", err)
	}
	return nil
}

// RevokeAllForUser revoca todas las sesiones de un usuario.
func (r *RefreshTokenRepository) RevokeAllForUser(userID string) error {
	now := time.Now().UTC()
	sql := `UPDATE refresh_tokens SET revoked_at = $1 WHERE user_id = $2 AND revoked_at IS NULL`
	_, err := r.pool.Exec(context.Background(), sql, now, userID)
	return err
}
