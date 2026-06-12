# RUNBOOK — Arranque del sistema

## Requisitos previos
- Go 1.25+
- PostgreSQL corriendo y accesible
- Variable de entorno DATABASE_URL configurada

## Pasos

1. Configurar variable de entorno:

   export DATABASE_URL="postgres://localhost:5432/vuhmik_dev?sslmode=disable"

2. Aplicar migraciones (forward-only):

   migrate -path database/migrations -database "$DATABASE_URL" up

3. Arrancar el servidor:

   go run ./cmd/vuhmik-api/

4. Verificar arranque exitoso:

   Debe mostrar en stdout (JSON):
   {"level":"INFO","msg":"servidor iniciado","addr":":8080"}

## Fallo esperado: secreto ausente

Si DATABASE_URL no está configurada, el servidor se niega a arrancar
(fail-closed) con:

   error de configuración: variables de entorno requeridas ausentes: [DATABASE_URL]

Esto es el comportamiento correcto. No es un bug.
