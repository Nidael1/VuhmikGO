# VUHMIK — Guia de deploy en VPS Ubuntu

## Requisitos en la VPS
- Ubuntu 22.04 o 24.04
- Go 1.21+
- PostgreSQL 15
- Redis 7
- Nginx
- Certbot (Let's Encrypt)
- golang-migrate

## Primera vez (setup inicial)

### 1. Crear usuario del sistema
```bash
sudo useradd -m -s /bin/bash vuhmik
sudo mkdir -p /opt/vuhmik/backups
sudo chown -R vuhmik:vuhmik /opt/vuhmik
```

### 2. Clonar repositorio
```bash
sudo -u vuhmik git clone https://github.com/Nidael1/VuhmikGO.git /opt/vuhmik/repo
```

### 3. Configurar variables de entorno
```bash
sudo cp /opt/vuhmik/repo/deploy/.env.example /opt/vuhmik/.env
sudo nano /opt/vuhmik/.env  # llenar con valores reales
```

### 4. Crear base de datos
```bash
sudo -u postgres psql -c "CREATE USER vuhmik_user WITH PASSWORD 'CONTRASENA_SEGURA';"
sudo -u postgres psql -c "CREATE DATABASE vuhmik_prod OWNER vuhmik_user;"
```

### 5. Primer build y migraciones
```bash
cd /opt/vuhmik/repo
go build -o /opt/vuhmik/vuhmik-api ./cmd/vuhmik-api/main.go
migrate -path database/migrations -database "$DATABASE_URL" up
```

### 6. Instalar servicio systemd
```bash
sudo cp /opt/vuhmik/repo/deploy/vuhmik.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable vuhmik
sudo systemctl start vuhmik
```

### 7. Configurar Nginx
```bash
# Reemplazar TU_DOMINIO.COM con tu dominio real
sudo cp /opt/vuhmik/repo/deploy/nginx.conf /etc/nginx/sites-available/vuhmik
sudo ln -s /etc/nginx/sites-available/vuhmik /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx
```

### 8. SSL con Certbot
```bash
sudo certbot --nginx -d TU_DOMINIO.COM -d www.TU_DOMINIO.COM
```

## Deploys posteriores
```bash
cd /opt/vuhmik/repo
./deploy/deploy.sh
```

## Generar JWT_SECRET seguro
```bash
openssl rand -hex 64
```
