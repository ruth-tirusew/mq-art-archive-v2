package config

import (
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DatabaseURL        string
	HTTPPort           int
	JWTSecret          string
	JWTAccessTTL       time.Duration
	CORSOrigins        []string
	GoogleClientID     string
	GoogleClientSecret string
	OAuthCallbackURL   string
	AuthCookieName     string
	AuthDevMode        bool
}

func Load() Config {
	port, _ := strconv.Atoi(getEnv("HTTP_PORT", "8080"))
	return Config{
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://mq:mq@localhost:5432/mq?sslmode=disable"),
		HTTPPort:           port,
		JWTSecret:          getEnv("JWT_SECRET", "change-me-in-production"),
		JWTAccessTTL:       parseDuration(getEnv("JWT_ACCESS_TTL", "1h"), time.Hour),
		CORSOrigins:        splitCSV(getEnv("CORS_ORIGINS", "http://localhost:5173,http://localhost:5174")),
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		OAuthCallbackURL:   getEnv("OAUTH_CALLBACK_URL", "http://localhost:8080/api/v1/auth/google/callback"),
		AuthCookieName:     getEnv("AUTH_COOKIE_NAME", "mq_access_token"),
		AuthDevMode:        getEnv("AUTH_DEV_MODE", "true") == "true",
	}
}

func parseDuration(raw string, fallback time.Duration) time.Duration {
	d, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}
	return d
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
