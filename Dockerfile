# Etapa 1: compilacion
FROM golang:1.26-alpine AS builder

WORKDIR /app

# go.mod y go.sum primero: su cache solo se invalida al cambiar dependencias
COPY go.mod go.sum ./
RUN go mod download

# Codigo fuente despues: cambia en cada commit
COPY . .

# Binario estatico sin debug symbols
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o vuhmik-api ./cmd/vuhmik-api/

# Etapa 2: imagen final minima
FROM alpine:3.21

# ca-certificates: TLS saliente para backups y webhooks
# tzdata: America/Mexico_City en logs
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/vuhmik-api .

EXPOSE 8080

RUN addgroup -S vuhmik && adduser -S vuhmik -G vuhmik
USER vuhmik

ENTRYPOINT ["./vuhmik-api"]
