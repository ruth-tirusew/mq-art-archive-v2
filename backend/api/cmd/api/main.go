package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/mq/api/config"
	"github.com/mq/api/internal/adapter/driven/auth"
	errormonitoradapter "github.com/mq/api/internal/adapter/driven/errormonitor"
	eventsadapter "github.com/mq/api/internal/adapter/driven/events"
	maileradapter "github.com/mq/api/internal/adapter/driven/mailer"
	mediaadapter "github.com/mq/api/internal/adapter/driven/media"
	oauthadapter "github.com/mq/api/internal/adapter/driven/oauth"
	"github.com/mq/api/internal/adapter/driven/persistence/postgres"
	telegramadapter "github.com/mq/api/internal/adapter/driven/telegram"
	httpadapter "github.com/mq/api/internal/adapter/driving/http"
	"github.com/mq/api/internal/adapter/driving/http/handler"
	"github.com/mq/api/internal/adapter/driving/http/middleware"
	"github.com/mq/api/internal/port/inbound"
	"github.com/mq/api/internal/port/outbound"
	analyticsuc "github.com/mq/api/internal/usecase/analytics"
	artuc "github.com/mq/api/internal/usecase/art"
	authuc "github.com/mq/api/internal/usecase/auth"
	contentuc "github.com/mq/api/internal/usecase/content"
	digestuc "github.com/mq/api/internal/usecase/digest"
	eventsuc "github.com/mq/api/internal/usecase/events"
	identityuc "github.com/mq/api/internal/usecase/identity"
	institutionuc "github.com/mq/api/internal/usecase/institution"
	mediauc "github.com/mq/api/internal/usecase/media"
	onboardinguc "github.com/mq/api/internal/usecase/onboarding"
	profileuc "github.com/mq/api/internal/usecase/profile"
	searchuc "github.com/mq/api/internal/usecase/search"
	settingsuc "github.com/mq/api/internal/usecase/settings"
	wikiuc "github.com/mq/api/internal/usecase/wiki"
)

func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("configuration: %v", err)
	}

	if err := postgres.Migrate(cfg.DatabaseURL); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := postgres.NewPool(ctx, cfg)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	// Repositories (driven adapters)
	articleRepo := postgres.NewArticleRepository(pool)
	artPostRepo := postgres.NewArtPostRepository(pool)
	profileRepo := postgres.NewProfileRepository(pool)
	institutionRepo := postgres.NewInstitutionRepository(pool)
	onboardingRepo := postgres.NewOnboardingRepository(pool)
	userRepo := postgres.NewUserRepository(pool)
	oauthAccountRepo := postgres.NewOAuthAccountRepository(pool)
	passwordResetRepo := postgres.NewPasswordResetRepository(pool)
	emailVerificationRepo := postgres.NewEmailVerificationRepository(pool)
	notifPrefsRepo := postgres.NewNotificationPreferencesRepository(pool)
	scrapeSettingsRepo := postgres.NewScrapeSettingsRepository(pool)
	eventRepo := postgres.NewEventRepository(pool)
	eventLocationRepo := postgres.NewEventLocationRepository(pool)
	mediaAssetRepo := postgres.NewMediaAssetRepository(pool)
	wikiSubmissionRepo := postgres.NewWikiSubmissionRepository(pool)
	analyticsRepo := postgres.NewAnalyticsRepository(pool)
	digestRecipientRepo := postgres.NewDigestRecipientRepository(pool)
	digestRunRepo := postgres.NewDigestRunRepository(pool)
	telegramLinkRepo := postgres.NewTelegramLinkRepository(pool)

	if err := settingsuc.EnsureSeed(ctx, scrapeSettingsRepo, eventsadapter.SettingsFromConfig(cfg)); err != nil {
		log.Fatalf("scrape settings seed: %v", err)
	}
	scrapeCfg, err := scrapeSettingsRepo.Get(ctx)
	if err != nil {
		log.Fatalf("scrape settings: %v", err)
	}
	eventSource := eventsadapter.NewSwappableEventSource(
		eventsadapter.NewEventSourceFromSettings(*scrapeCfg, cfg.TelegramSessionPath),
	)
	eventReloader := eventsadapter.NewEventSourceReloader(eventSource, cfg.TelegramSessionPath)

	// Auth adapters
	tokenSvc := auth.NewTokenService(cfg.JWTSecret, cfg.JWTAccessTTL)
	stateSvc := auth.NewOAuthStateService(cfg.JWTSecret, 10*time.Minute)

	var oauthProviders []outbound.OAuthProvider
	if cfg.GoogleClientID != "" && cfg.GoogleClientSecret != "" {
		oauthProviders = append(oauthProviders, oauthadapter.NewGoogleProvider(
			cfg.GoogleClientID,
			cfg.GoogleClientSecret,
			cfg.OAuthCallbackURL,
		))
	}

	// Use cases
	contentSvc := contentuc.NewService(articleRepo)
	artSvc := artuc.NewService(artPostRepo)
	profileSvc := profileuc.NewService(profileRepo, userRepo, artPostRepo)
	institutionSvc := institutionuc.NewService(institutionRepo)
	onboardingSvc := onboardinguc.NewService(onboardingRepo, userRepo, profileRepo, institutionRepo)
	searchSvc := searchuc.NewService(articleRepo, eventRepo, profileRepo, artPostRepo)
	identitySvc := identityuc.NewService(userRepo)
	settingsSvc := settingsuc.NewService(scrapeSettingsRepo, eventReloader, cfg.TelegramSessionPath)
	cloudinary := mediaadapter.NewCloudinary(cfg.CloudinaryCloudName, cfg.CloudinaryAPIKey, cfg.CloudinaryAPISecret, cfg.CloudinaryFolder, 5*time.Minute)
	mediaSvc := mediauc.NewService(cloudinary, cloudinary, mediaAssetRepo, cfg.CloudinaryFolder)
	wikiSvc := wikiuc.NewService(wikiSubmissionRepo, contentSvc)
	analyticsSvc := analyticsuc.NewService(analyticsRepo)
	passwordHasher := auth.NewBcryptPasswordHasher()
	var mailer outbound.Mailer = maileradapter.NewLogMailer()
	if cfg.ResendAPIKey != "" {
		mailer = maileradapter.NewResendMailer(cfg.ResendAPIKey, cfg.MailFrom)
	}
	// telegramBot is kept as its concrete type (not just outbound.TelegramNotifier) so the
	// polling goroutine below can reuse this same instance rather than constructing a
	// second one. telegramNotifier stays a nil interface when the bot isn't configured,
	// not a non-nil interface wrapping a nil pointer — digest.Service checks it with a
	// plain != nil, which only works correctly if unset means a truly nil interface.
	var telegramBot *telegramadapter.Bot
	var telegramNotifier outbound.TelegramNotifier
	if cfg.TelegramBotToken != "" {
		telegramBot = telegramadapter.NewBot(cfg.TelegramBotToken)
		telegramNotifier = telegramBot
	}
	eventsSvc := eventsuc.NewService(eventRepo, eventLocationRepo, eventSource, notifPrefsRepo, mailer)
	digestSvc := digestuc.NewService(
		articleRepo, eventRepo, artPostRepo, digestRecipientRepo, digestRunRepo,
		notifPrefsRepo, telegramLinkRepo, cfg.JWTSecret,
		mailer, telegramNotifier, cfg.WebAppURL, cfg.PublicAPIURL,
	)
	authSvc := authuc.NewService(
		userRepo,
		oauthAccountRepo,
		tokenSvc,
		passwordHasher,
		stateSvc,
		oauthProviders,
		cfg.CORSOrigins,
		mailer,
		passwordResetRepo,
		notifPrefsRepo,
		cfg.WebAppURL,
		emailVerificationRepo,
	)

	authCfg := middleware.AuthConfig{
		Verifier:   tokenSvc,
		Identity:   identitySvc,
		CookieName: cfg.AuthCookieName,
		DevMode:    cfg.AuthDevMode,
	}
	var errorMonitor outbound.ErrorMonitor = errormonitoradapter.NewNoop()
	if cfg.ErrorMonitorDSN != "" {
		errorMonitor = errormonitoradapter.NewLog(cfg.ErrorMonitorDSN)
	}

	// HTTP handlers (driving adapters)
	handlers := httpadapter.Handlers{
		Health:      handler.NewHealthHandler(),
		Article:     handler.NewArticleHandler(contentSvc),
		Profile:     handler.NewProfileHandler(profileSvc),
		Institution: handler.NewInstitutionHandler(institutionSvc),
		Art:         handler.NewArtHandler(artSvc, profileSvc),
		Event:       handler.NewEventHandler(eventsSvc),
		Search:      handler.NewSearchHandler(searchSvc),
		Onboarding:  handler.NewOnboardingHandler(onboardingSvc),
		Auth:        handler.NewAuthHandler(authSvc, cfg.AuthCookieName),
		Settings:    handler.NewSettingsHandler(settingsSvc),
		Media:       handler.NewMediaHandler(mediaSvc),
		Wiki:        handler.NewWikiHandler(wikiSvc),
		Analytics:   handler.NewAnalyticsHandler(analyticsSvc),
		UserAdmin:   handler.NewUserAdminHandler(authSvc),
		Digest:      handler.NewDigestHandler(digestSvc, cfg.TelegramBotUsername),
	}

	router := httpadapter.NewRouter(cfg, handlers, httpadapter.RouterDeps{
		Identity: identitySvc,
		Auth:     authCfg,
		Monitor:  errorMonitor,
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler: router,
	}

	go func() {
		log.Printf("api listening on :%d", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	botCtx, botCancel := context.WithCancel(context.Background())
	defer botCancel()
	if telegramBot != nil {
		go func() {
			log.Printf("telegram bot polling for updates")
			if err := telegramBot.PollUpdates(botCtx, telegramUpdateHandler(digestSvc, telegramBot)); err != nil && botCtx.Err() == nil {
				log.Printf("telegram bot polling stopped: %v", err)
			}
		}()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	botCancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
}

// telegramUpdateHandler wires incoming bot messages to the digest usecase's link/unlink
// flow. It's the only place that knows the bot's text commands (/start <token>, /stop);
// everything else just deals in chat IDs and tokens.
func telegramUpdateHandler(digestSvc inbound.DigestService, bot *telegramadapter.Bot) telegramadapter.Handler {
	return func(ctx context.Context, chatID, text string) {
		switch {
		case strings.HasPrefix(text, "/start"):
			if err := digestSvc.HandleTelegramStart(ctx, chatID, text); err != nil {
				log.Printf("telegram link error: %v", err)
				_ = bot.Send(ctx, chatID, "That link has expired or was already used. Generate a new one from your account settings.")
				return
			}
			_ = bot.Send(ctx, chatID, "You're linked. You'll get the weekly Artiv digest here — send /stop anytime to unsubscribe.")
		case strings.HasPrefix(text, "/stop"):
			if err := digestSvc.HandleTelegramStop(ctx, chatID); err != nil {
				log.Printf("telegram unlink error: %v", err)
				return
			}
			_ = bot.Send(ctx, chatID, "You've been unsubscribed from the Artiv digest.")
		}
	}
}
