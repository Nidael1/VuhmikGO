# Guía de arranque local — VUHMÍK

Para correr el proyecto en cualquier máquina desde cero.  
Incluye instrucciones para **Mac**, **Windows** y **Linux**.

---

## Requisitos

| Herramienta | Versión | Mac | Windows | Linux (Ubuntu/Debian) |
|---|---|---|---|---|
| Go | 1.25+ | `brew install go` | [go.dev/dl](https://go.dev/dl) → installer | `sudo snap install go --classic` |
| Node.js | 22 | NVM (ver abajo) | NVM for Windows (ver abajo) | NVM (ver abajo) |
| PostgreSQL | 15+ | `brew install postgresql@15` | [postgresql.org/download/windows](https://www.postgresql.org/download/windows/) | `sudo apt install postgresql` |
| Redis | 7+ | `brew install redis` | [github.com/tporadowski/redis/releases](https://github.com/tporadowski/redis/releases) → .msi | `sudo apt install redis-server` |
| golang-migrate | 4.18+ | `brew install golang-migrate` | Binario en releases | `curl` (ver abajo) |
| Git | cualquiera | `brew install git` | [git-scm.com](https://git-scm.com/download/win) | `sudo apt install git` |

---

## 1. Clonar el repositorio

```bash
git clone https://github.com/Nidael1/VuhmikGO.git
cd VuhmikGO
```

---

## 2. Variables de entorno

```bash
# Mac / Linux
cp .env.example .env
nano .env   # o el editor que prefieras

# Windows (PowerShell)
Copy-Item .env.example .env
notepad .env
```

**Solo necesitas cambiar `tu_password`** — el resto ya está configurado (`vuhmik` como usuario y nombre de BD).

---

## 3. Base de datos (PostgreSQL)

### Mac

```bash
brew services start postgresql@15

psql postgres -c "CREATE USER vuhmik WITH PASSWORD 'tu_password';"
psql postgres -c "CREATE DATABASE vuhmik OWNER vuhmik;"

export $(cat .env | grep -v '#' | xargs)
migrate -path database/migrations -database "$DATABASE_URL" up
```

### Linux (Ubuntu/Debian)

```bash
sudo systemctl start postgresql
sudo systemctl enable postgresql

# Crear usuario y base de datos
sudo -u postgres psql -c "CREATE USER vuhmik WITH PASSWORD 'tu_password';"
sudo -u postgres psql -c "CREATE DATABASE vuhmik OWNER vuhmik;"

# Instalar golang-migrate
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.1/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/

# Correr migraciones
export $(cat .env | grep -v '#' | xargs)
migrate -path database/migrations -database "$DATABASE_URL" up
```

### Windows (PowerShell)

```powershell
# PostgreSQL se instala como servicio — ya corre automáticamente.
# Usar pgAdmin o psql:

psql -U postgres -c "CREATE USER vuhmik WITH PASSWORD 'tu_password';"
psql -U postgres -c "CREATE DATABASE vuhmik OWNER vuhmik;"

# golang-migrate: descargar migrate.windows-amd64.zip de:
# https://github.com/golang-migrate/migrate/releases
# Extraer migrate.exe y agregar al PATH (ej. C:\tools\)

$env:DATABASE_URL = "postgres://vuhmik:tu_password@localhost:5432/vuhmik?sslmode=disable"
migrate.exe -path database/migrations -database $env:DATABASE_URL up
```

---

## 4. Redis

### Mac

```bash
brew services start redis
```

### Linux (Ubuntu/Debian)

```bash
sudo systemctl start redis-server
sudo systemctl enable redis-server

# Verificar
redis-cli ping   # → PONG
```

### Windows

Instalar con el `.msi` de [github.com/tporadowski/redis/releases](https://github.com/tporadowski/redis/releases).  
Redis se instala como servicio y arranca automáticamente.  
Verificar: `redis-cli ping` → `PONG`.

---

## 5. Backend (Go)

### Mac / Linux

```bash
export $(cat .env | grep -v '#' | xargs)
go run ./cmd/vuhmik-api/
# → http://localhost:8080
```

### Windows (PowerShell)

```powershell
# Cargar variables de entorno desde .env
Get-Content .env | Where-Object { $_ -notmatch '^#' -and $_ -ne '' } | ForEach-Object {
    $parts = $_ -split '=', 2
    [System.Environment]::SetEnvironmentVariable($parts[0], $parts[1], 'Process')
}

go run ./cmd/vuhmik-api/
# → http://localhost:8080
```

> **Nota Windows:** crear los directorios antes de arrancar:
> ```powershell
> mkdir C:\Users\$env:USERNAME\vuhmik-backups
> mkdir C:\Users\$env:USERNAME\vuhmik-logs
> ```

> **Nota Linux:** crear los directorios si no existen:
> ```bash
> mkdir -p /tmp/vuhmik-backups /tmp/vuhmik-logs
> ```

---

## 6. Frontend (Vue 3 + Vite)

### Mac — NVM en `/Volumes/D/nvm`

```bash
export NVM_DIR="/Volumes/D/nvm"
source "$NVM_DIR/nvm.sh"
nvm use 22

cd frontend
npm install
npm run dev
# → http://localhost:5173
```

### Linux — NVM estándar

```bash
# Instalar NVM (si no está instalado)
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash
source ~/.bashrc   # o ~/.zshrc si usas zsh

nvm install 22
nvm use 22

cd frontend
npm install
npm run dev
# → http://localhost:5173
```

### Windows — NVM for Windows

```powershell
# Instalar desde: https://github.com/coreybutler/nvm-windows/releases → nvm-setup.exe

nvm install 22
nvm use 22

cd frontend
npm install
npm run dev
# → http://localhost:5173
```

---

## 7. Verificar que todo funciona

```bash
# Backend
curl http://localhost:8080/api/v1/health

# Frontend
# Abrir en el navegador: http://localhost:5173
```

> Usar **siempre** `localhost:5173` en desarrollo, no `:8080` directamente.  
> Vite hace proxy automático al backend.

---

## Producción (VPS Hetzner via Coolify)

El despliegue es automático via Docker — no se hace nada manual.

```bash
# Build manual solo para probar la imagen:
docker build -t vuhmik .
docker run -p 8080:8080 --env-file .env vuhmik
```

Las variables de entorno de producción se configuran en el panel de Coolify — **nunca en el repositorio**.

---

## Resumen de puertos

| Servicio | Puerto |
|---|---|
| API (Go) | 8080 |
| Frontend (Vite dev) | 5173 |
| PostgreSQL | 5432 |
| Redis | 6379 |

---

*NDT — Next Dev Tech. Actualizado 2026-08-21.*
