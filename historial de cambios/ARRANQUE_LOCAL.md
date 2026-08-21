# Guía de arranque local — VUHMÍK

Para poder correr el proyecto en cualquier máquina desde cero.

---

## Requisitos

| Herramienta | Versión | Notas |
|---|---|---|
| Go | 1.25+ | `brew install go` |
| Node.js | 22 | Usar NVM (ver abajo) |
| PostgreSQL | 15+ | `brew install postgresql` |
| Redis | 7+ | `brew install redis` |
| golang-migrate | 4.18+ | `brew install golang-migrate` |

---

## 1. Clonar el repositorio

```bash
git clone https://github.com/Nidael1/VuhmikGO.git
cd VuhmikGO
```

---

## 2. Variables de entorno

```bash
cp .env.example .env
# Edita .env con tus valores reales
```

Variables requeridas:
- `DATABASE_URL` — PostgreSQL connection string
- `REDIS_URL` — Redis URL (redis://localhost:6379 en dev)
- `JWT_SECRET` — Secreto JWT (mín. 32 chars). Genera con: `openssl rand -hex 32`
- `BACKUP_DIR` — Ruta de backups (ej. `/tmp/vuhmik-backups`)
- `LOG_DIR` — Ruta de logs (ej. `/tmp/vuhmik-logs`)

---

## 3. Base de datos

```bash
# Crear la base de datos
createdb vuhmik

# Correr migraciones
migrate -path database/migrations -database "$DATABASE_URL" up
```

---

## 4. Redis

```bash
redis-server
```

---

## 5. Backend (Go)

```bash
# Cargar variables de entorno
export $(cat .env | grep -v '#' | xargs)

# Correr el servidor
go run ./cmd/vuhmik-api/
# → escucha en :8080
```

---

## 6. Frontend (Vue 3 + Vite)

```bash
# NVM — cargar Node 22
export NVM_DIR="/Volumes/D/nvm"   # Mac local NDT
# En otra máquina: export NVM_DIR="$HOME/.nvm"
source "$NVM_DIR/nvm.sh"
nvm use 22

cd frontend
npm install
npm run dev
# → escucha en http://localhost:5173
```

> La app en desarrollo apunta al backend en `:8080` vía proxy Vite.
> Usar **siempre** `localhost:5173` para desarrollo, no `:8080` directamente.

---

## 7. Verificar que todo funciona

```bash
# Backend health
curl http://localhost:8080/api/v1/health

# Frontend
open http://localhost:5173
```

---

## Producción (VPS Hetzner via Coolify)

El despliegue en producción es automático via Docker:

```bash
# Build manual (solo si es necesario probar la imagen)
docker build -t vuhmik .
docker run -p 8080:8080 --env-file .env vuhmik
```

El `Dockerfile` multi-etapa:
1. Compila el frontend con Node 22.
2. Compila el binario Go.
3. Descarga `golang-migrate` para Linux amd64.
4. Imagen final Alpine mínima con el binario, migraciones y frontend compilado.

El `docker-entrypoint.sh` corre las migraciones automáticamente antes de arrancar.

Las variables de entorno se configuran directamente en Coolify — **no en el repositorio**.

---

## .gitignore recomendado

```
.env
*.env.local
dump.rdb
```

---

*NDT — Next Dev Tech. Actualizado 2026-08-21.*
