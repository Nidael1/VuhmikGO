// Package auth implementa autenticacion JWT para la API post-MVP.
//
// Responsabilidades:
//   - Hash y verificacion de passwords con bcrypt.
//   - Generacion y validacion de JWT firmados con HMAC-SHA256.
//   - El token contiene tenant_id y actor_id resueltos por el servidor.
//   - El cliente nunca decide quien es el tenant.
//
// No contiene logica clinica ni reglas del Core.
package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Claims contiene los campos del JWT.
type Claims struct {
	TenantID string `json:"tenant_id"`
	ActorID  string `json:"actor_id"`
	jwt.RegisteredClaims
}

// HashPassword genera un hash bcrypt del password dado.
// Nunca almacenar passwords en texto plano.
func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// CheckPassword verifica que el password coincide con el hash almacenado.
func CheckPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// GenerateToken genera un JWT firmado para el usuario dado.
// Expira en 24 horas. Usa JWT_SECRET del entorno.
func GenerateToken(userID, tenantID string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", errors.New("JWT_SECRET no configurado")
	}
	claims := Claims{
		TenantID: tenantID,
		ActorID:  userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateToken valida un JWT y retorna los Claims si es valido.
func ValidateToken(tokenStr string) (*Claims, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, errors.New("JWT_SECRET no configurado")
	}
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("metodo de firma invalido")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("token invalido")
	}
	return claims, nil
}
