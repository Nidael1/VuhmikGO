// Package redis provee el cliente Redis para VUHMÍK.
// Redis se usa como broker de jobs, locks y cache efimero.
// NUNCA se almacena evidencia clínica en Redis.
// Redis es efimero — no es fuente de verdad, solo acelerador.
package redis

import (
	"context"
	"fmt"
	"os"

	goredis "github.com/redis/go-redis/v9"
)

// Client es el cliente Redis de VUHMÍK.
type Client struct {
	rdb *goredis.Client
}

// NewClient crea un cliente Redis desde la variable de entorno REDIS_URL.
// Formato: redis://localhost:6379
func NewClient() (*Client, error) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}
	opt, err := goredis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("error al parsear REDIS_URL: %w", err)
	}
	rdb := goredis.NewClient(opt)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("error al conectar Redis: %w", err)
	}
	return &Client{rdb: rdb}, nil
}

// Close cierra la conexion Redis.
func (c *Client) Close() error {
	return c.rdb.Close()
}

// RDB retorna el cliente raw de go-redis para uso en workers.
func (c *Client) RDB() *goredis.Client {
	return c.rdb
}

// Ping verifica la conexion Redis.
func (c *Client) Ping() error {
	return c.rdb.Ping(context.Background()).Err()
}
