package config

import (
	"testing"

	"github.com/mq/api/internal/testutil/assist"
)

func TestLoad_defaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("HTTP_PORT", "")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("CORS_ORIGINS", "")

	cfg := Load()
	assist.Equal(t, "postgres://mq:mq@localhost:5432/mq?sslmode=disable", cfg.DatabaseURL)
	assist.Equal(t, 8080, cfg.HTTPPort)
	assist.Equal(t, "change-me-in-production", cfg.JWTSecret)
	assist.Equal(t, 2, len(cfg.CORSOrigins))
}

func TestLoad_fromEnv(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://custom/db")
	t.Setenv("HTTP_PORT", "9090")
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("CORS_ORIGINS", "http://a.test,http://b.test")

	cfg := Load()
	assist.Equal(t, "postgres://custom/db", cfg.DatabaseURL)
	assist.Equal(t, 9090, cfg.HTTPPort)
	assist.Equal(t, "secret", cfg.JWTSecret)
	assist.Equal(t, "http://a.test", cfg.CORSOrigins[0])
}
