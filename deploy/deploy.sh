#!/bin/bash
# VUHMIK — Script de deploy en VPS
# Uso: ./deploy/deploy.sh
# Ejecutar en la VPS como usuario vuhmik

set -e

APP_DIR=/opt/vuhmik
BACKUP_DIR=/opt/vuhmik/backups
REPO_DIR=/opt/vuhmik/repo

echo "[deploy] Iniciando deploy de VUHMIK..."

# 1. Actualizar codigo
cd $REPO_DIR
git pull origin main
echo "[deploy] Codigo actualizado"

# 2. Compilar backend
cd $REPO_DIR
go build -o $APP_DIR/vuhmik-api ./cmd/vuhmik-api/main.go
echo "[deploy] Backend compilado"

# 3. Compilar frontend
cd $REPO_DIR/frontend
npm ci --silent
npm run build
cp -r dist $APP_DIR/frontend/dist
echo "[deploy] Frontend compilado"

# 4. Aplicar migraciones pendientes
/usr/local/bin/migrate     -path $REPO_DIR/database/migrations     -database "$DATABASE_URL"     up
echo "[deploy] Migraciones aplicadas"

# 5. Reiniciar servicio
sudo systemctl restart vuhmik
sleep 2
sudo systemctl status vuhmik --no-pager
echo "[deploy] Servicio reiniciado"

echo "[deploy] Deploy completado exitosamente"
