// Package config reads process configuration from the environment.
package config

import (
	"fmt"
	"os"
	"strings"
)

// Config is everything the process needs to boot.
//
// The localhost defaults match docker-compose.ports.yml (5433 and 6380),
// not 5432 and 6379. Those two are often already taken on a dev machine.
// Compose itself does not use these defaults: the backend container sets
// DATABASE_URL and REDIS_URL to the hostnames db and redis.
type Config struct {
	DatabaseURL   string
	RedisURL      string
	HTTPAddr      string
	CORSOrigins   []string
	MigrationsDir string
}

// Load reads env vars. An empty DATABASE_URL or REDIS_URL is a mistake,
// not a signal to guess a production host.
func Load() (Config, error) {
	cfg := Config{
		DatabaseURL:   getenv("DATABASE_URL", "postgres://logistic:logistic@127.0.0.1:5433/logistic?sslmode=disable"),
		RedisURL:      getenv("REDIS_URL", "redis://127.0.0.1:6380/0"),
		HTTPAddr:      getenv("HTTP_ADDR", ":8080"),
		MigrationsDir: getenv("MIGRATIONS_DIR", "migrations"),
	}
	rawOrigins := getenv("CORS_ORIGINS", "http://localhost:3000,http://127.0.0.1:3000")
	for _, origin := range strings.Split(rawOrigins, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			cfg.CORSOrigins = append(cfg.CORSOrigins, origin)
		}
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.RedisURL == "" {
		return Config{}, fmt.Errorf("REDIS_URL is required")
	}
	if len(cfg.CORSOrigins) == 0 {
		return Config{}, fmt.Errorf("CORS_ORIGINS must list at least one origin")
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
