package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv              string
	DatabaseURL         string
	HTTPPort            int
	JWTSecret           string
	JWTAccessTTL        time.Duration
	CORSOrigins         []string
	GoogleClientID      string
	GoogleClientSecret  string
	OAuthCallbackURL    string
	AuthCookieName      string
	AuthDevMode         bool
	ResendAPIKey        string
	MailFrom            string
	WebAppURL           string
	ErrorMonitorDSN     string
	CloudinaryEnabled   bool
	CloudinaryCloudName string
	CloudinaryAPIKey    string
	CloudinaryAPISecret string
	CloudinaryFolder    string

	// Scrape / RSS
	ScrapeEnabled   bool
	ScrapeSources   []string
	ScrapeUserAgent string
	ScrapeTimeout   time.Duration
	ScrapeInterval  time.Duration

	// Telegram MTProto
	TelegramEnabled     bool
	TelegramAPIID       int
	TelegramAPIHash     string
	TelegramSessionPath string
	TelegramChannels    []string
	TelegramKeywords    []string
	TelegramFetchLimit  int
}

func Load() Config {
	port, _ := strconv.Atoi(getEnv("HTTP_PORT", "8080"))
	apiID, _ := strconv.Atoi(getEnv("TELEGRAM_API_ID", "0"))
	fetchLimit, _ := strconv.Atoi(getEnv("TELEGRAM_FETCH_LIMIT", "50"))
	if fetchLimit <= 0 {
		fetchLimit = 50
	}

	webAppURL := getEnv("WEB_APP_URL", "http://localhost:5173")
	corsOrigins := mergeCORSOrigins(splitCSV(getEnv("CORS_ORIGINS", "http://localhost:5173,http://localhost:5174")), webAppURL)

	return Config{
		AppEnv:              getEnv("APP_ENV", "development"),
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://mq:mq@localhost:5432/mq?sslmode=disable"),
		HTTPPort:            port,
		JWTSecret:           getEnv("JWT_SECRET", "change-me-in-production"),
		JWTAccessTTL:        parseDuration(getEnv("JWT_ACCESS_TTL", "1h"), time.Hour),
		CORSOrigins:         corsOrigins,
		GoogleClientID:      getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:  getEnv("GOOGLE_CLIENT_SECRET", ""),
		OAuthCallbackURL:    getEnv("OAUTH_CALLBACK_URL", "http://localhost:8080/api/v1/auth/google/callback"),
		AuthCookieName:      getEnv("AUTH_COOKIE_NAME", "mq_access_token"),
		AuthDevMode:         getEnv("AUTH_DEV_MODE", "true") == "true",
		ResendAPIKey:        getEnv("RESEND_API_KEY", ""),
		MailFrom:            getEnv("MAIL_FROM", "Artiv <noreply@artiv.local>"),
		WebAppURL:           webAppURL,
		ErrorMonitorDSN:     getEnv("ERROR_MONITOR_DSN", ""),
		CloudinaryEnabled:   getEnv("CLOUDINARY_ENABLED", "false") == "true",
		CloudinaryCloudName: getEnv("CLOUDINARY_CLOUD_NAME", ""),
		CloudinaryAPIKey:    getEnv("CLOUDINARY_API_KEY", ""),
		CloudinaryAPISecret: getEnv("CLOUDINARY_API_SECRET", ""),
		CloudinaryFolder:    getEnv("CLOUDINARY_FOLDER", ""),

		ScrapeEnabled:   getEnv("SCRAPE_ENABLED", "false") == "true",
		ScrapeSources:   splitCSV(getEnv("SCRAPE_SOURCES", "")),
		ScrapeUserAgent: getEnv("SCRAPE_USER_AGENT", "mq-scraper/1.0"),
		ScrapeTimeout:   parseDuration(getEnv("SCRAPE_TIMEOUT", "30s"), 30*time.Second),
		ScrapeInterval:  parseDuration(getEnv("SCRAPE_INTERVAL", "6h"), 6*time.Hour),

		TelegramEnabled:     getEnv("TELEGRAM_ENABLED", "false") == "true",
		TelegramAPIID:       apiID,
		TelegramAPIHash:     getEnv("TELEGRAM_API_HASH", ""),
		TelegramSessionPath: getEnv("TELEGRAM_SESSION_PATH", "data/telegram.session"),
		TelegramChannels:    splitCSV(getEnv("TELEGRAM_CHANNELS", "")),
		TelegramKeywords:    splitCSV(getEnv("TELEGRAM_KEYWORDS", "")),
		TelegramFetchLimit:  fetchLimit,
	}
}

func (c Config) Validate() error {
	if strings.ToLower(c.AppEnv) != "production" {
		return nil
	}
	if strings.TrimSpace(c.DatabaseURL) == "" {
		return fmt.Errorf("DATABASE_URL is required in production")
	}
	secret := strings.TrimSpace(c.JWTSecret)
	if secret == "" || strings.Contains(strings.ToLower(secret), "change-me") {
		return fmt.Errorf("JWT_SECRET must be changed in production")
	}
	for _, origin := range c.CORSOrigins {
		lower := strings.ToLower(origin)
		if origin == "*" || strings.Contains(lower, "localhost") || strings.Contains(lower, "127.0.0.1") {
			return fmt.Errorf("unsafe CORS origin %q in production", origin)
		}
	}
	if c.CloudinaryEnabled {
		if c.CloudinaryCloudName == "" || c.CloudinaryAPIKey == "" || c.CloudinaryAPISecret == "" || c.CloudinaryFolder == "" {
			return fmt.Errorf("all CLOUDINARY_* settings are required when enabled in production")
		}
	}
	if c.AuthDevMode {
		return fmt.Errorf("AUTH_DEV_MODE must be false in production")
	}
	return nil
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

// mergeCORSOrigins ensures the public web app origin can call the API with credentials.
func mergeCORSOrigins(origins []string, webAppURL string) []string {
	webAppURL = strings.TrimSpace(webAppURL)
	if webAppURL == "" {
		return origins
	}
	for _, origin := range origins {
		if origin == webAppURL {
			return origins
		}
	}
	return append(origins, webAppURL)
}
