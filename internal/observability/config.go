package observability

import (
	"fmt"
	"os"
)

// requiredEnvVars lista las variables de entorno obligatorias al arranque.
// La app falla inmediatamente si alguna falta.
var requiredEnvVars = []string{
	"DATABASE_URL",
	"JWT_SECRET",
	"REDIS_URL",
}

// ValidateRuntimeSecrets verifica que todas las variables de entorno
// requeridas estén presentes y no vacías.
//
// Reglas absolutas:
//   - No imprime valores de secretos en logs ni en errores.
//   - Solo reporta el NOMBRE de la variable faltante.
//   - La app debe fallar al arranque (fail-closed) si falta algún secreto.
func ValidateRuntimeSecrets() error {
	var missing []string
	for _, key := range requiredEnvVars {
		if os.Getenv(key) == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("variables de entorno requeridas ausentes: %v", missing)
	}
	return nil
}
