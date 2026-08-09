# Etapa 1: compilacion frontend
FROM node:22-alpine AS frontend-builder

WORKDIR /frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Etapa 2: compilacion backend
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o vuhmik-api ./cmd/vuhmik-api/

# Descargar golang-migrate para Linux amd64
RUN apk add --no-cache wget \
    && wget -q https://github.com/golang-migrate/migrate/releases/download/v4.18.1/migrate.linux-amd64.tar.gz \
    && tar xzf migrate.linux-amd64.tar.gz \
    && mv migrate /app/migrate \
    && rm migrate.linux-amd64.tar.gz \
    && apk del wget

# Etapa 3: imagen final minima
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata postgresql-client wget

WORKDIR /app

# Binario principal
COPY --from=builder /app/vuhmik-api .

# Herramienta de migraciones
COPY --from=builder /app/migrate .

# Archivos de migraciones
COPY database/migrations ./migrations

# Templates HTML requeridos en runtime por handlers.go
COPY internal/delivery/http/templates ./internal/delivery/http/templates

# Frontend compilado
COPY --from=frontend-builder /frontend/dist ./frontend/dist

# Script de entrada
COPY docker-entrypoint.sh .
RUN chmod +x docker-entrypoint.sh

EXPOSE 8080

RUN addgroup -S vuhmik \
    && adduser -S vuhmik -G vuhmik \
    && mkdir -p /app/logs \
    && chown -R vuhmik:vuhmik /app/logs

USER vuhmik

ENTRYPOINT ["./docker-entrypoint.sh"]
