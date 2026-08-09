#!/bin/sh
set -e

# Correr migraciones antes de arrancar el servidor.
# Si fallan, el contenedor no arranca y Coolify mantiene el despliegue anterior.
echo "[entrypoint] Corriendo migraciones..."
/app/migrate -path /app/migrations -database "$DATABASE_URL" -verbose up

echo "[entrypoint] Migraciones completadas. Arrancando servidor..."
exec /app/vuhmik-api
