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
	t.Setenv("SCRAPE_ENABLED", "true")
	t.Setenv("SCRAPE_SOURCES", "https://a.test/feed,https://b.test/rss")
	t.Setenv("TELEGRAM_ENABLED", "true")
	t.Setenv("TELEGRAM_API_ID", "12345")
	t.Setenv("TELEGRAM_CHANNELS", "addisart")

	cfg := Load()
	assist.Equal(t, "postgres://custom/db", cfg.DatabaseURL)
	assist.Equal(t, 9090, cfg.HTTPPort)
	assist.Equal(t, "secret", cfg.JWTSecret)
	assist.Equal(t, "http://a.test", cfg.CORSOrigins[0])
	assist.Equal(t, true, cfg.ScrapeEnabled)
	assist.Equal(t, 2, len(cfg.ScrapeSources))
	assist.Equal(t, true, cfg.TelegramEnabled)
	assist.Equal(t, 12345, cfg.TelegramAPIID)
	assist.Equal(t, "addisart", cfg.TelegramChannels[0])
}

func TestValidateProduction(t *testing.T) {
	valid := Config{
		AppEnv: "production", DatabaseURL: "postgres://db/prod", JWTSecret: "a-real-secret",
		CORSOrigins: []string{"https://artiv.example"}, CloudinaryEnabled: true,
		CloudinaryCloudName: "cloud", CloudinaryAPIKey: "key", CloudinaryAPISecret: "secret", CloudinaryFolder: "mq",
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid production config: %v", err)
	}
	cases := []Config{
		{AppEnv: "production", DatabaseURL: valid.DatabaseURL, JWTSecret: "change-me", CORSOrigins: valid.CORSOrigins},
		{AppEnv: "production", DatabaseURL: valid.DatabaseURL, JWTSecret: valid.JWTSecret, CORSOrigins: []string{"*"}},
		{AppEnv: "production", DatabaseURL: valid.DatabaseURL, JWTSecret: valid.JWTSecret, CORSOrigins: []string{"http://localhost:5173"}},
		{AppEnv: "production", JWTSecret: valid.JWTSecret, CORSOrigins: valid.CORSOrigins},
	}
	for i, cfg := range cases {
		if err := cfg.Validate(); err == nil {
			t.Errorf("case %d: expected validation error", i)
		}
	}
}
