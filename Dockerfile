# Etapa 1: compilacion
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

# Etapa 2: imagen final minima
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Binario principal
COPY --from=builder /app/vuhmik-api .

# Herramienta de migraciones
COPY --from=builder /app/migrate .

# Archivos de migraciones
COPY database/migrations ./migrations

# Script de entrada
COPY docker-entrypoint.sh .
RUN chmod +x docker-entrypoint.sh

EXPOSE 8080

RUN addgroup -S vuhmik && adduser -S vuhmik -G vuhmik
USER vuhmik

ENTRYPOINT ["./docker-entrypoint.sh"]
