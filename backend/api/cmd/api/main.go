package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mq/api/config"
	"github.com/mq/api/internal/adapter/driven/auth"
	eventsadapter "github.com/mq/api/internal/adapter/driven/events"
	oauthadapter "github.com/mq/api/internal/adapter/driven/oauth"
	"github.com/mq/api/internal/adapter/driven/persistence/postgres"
	httpadapter "github.com/mq/api/internal/adapter/driving/http"
	"github.com/mq/api/internal/adapter/driving/http/handler"
	"github.com/mq/api/internal/adapter/driving/http/middleware"
	"github.com/mq/api/internal/port/inbound"
	"github.com/mq/api/internal/port/outbound"
	artuc "github.com/mq/api/internal/usecase/art"
	authuc "github.com/mq/api/internal/usecase/auth"
	contentuc "github.com/mq/api/internal/usecase/content"
	eventsuc "github.com/mq/api/internal/usecase/events"
	identityuc "github.com/mq/api/internal/usecase/identity"
	institutionuc "github.com/mq/api/internal/usecase/institution"
	onboardinguc "github.com/mq/api/internal/usecase/onboarding"
	profileuc "github.com/mq/api/internal/usecase/profile"
)

func main() {
	cfg := config.Load()

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
	eventRepo := postgres.NewEventRepository(pool)
	eventLocationRepo := postgres.NewEventLocationRepository(pool)
	eventSource := eventsadapter.NewScraperNoop()

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
	profileSvc := profileuc.NewService(profileRepo)
	institutionSvc := institutionuc.NewService(institutionRepo)
	onboardingSvc := onboardinguc.NewService(onboardingRepo)
	eventsSvc := eventsuc.NewService(eventRepo, eventLocationRepo, eventSource)
	identitySvc := identityuc.NewService(userRepo)
	authSvc := authuc.NewService(userRepo, oauthAccountRepo, tokenSvc, stateSvc, oauthProviders, cfg.CORSOrigins)

	var _ inbound.EventsService = eventsSvc

	authCfg := middleware.AuthConfig{
		Verifier:   tokenSvc,
		Identity:   identitySvc,
		CookieName: cfg.AuthCookieName,
		DevMode:    cfg.AuthDevMode,
	}

	// HTTP handlers (driving adapters)
	handlers := httpadapter.Handlers{
		Health:      handler.NewHealthHandler(),
		Article:     handler.NewArticleHandler(contentSvc),
		Profile:     handler.NewProfileHandler(profileSvc),
		Institution: handler.NewInstitutionHandler(institutionSvc),
		Art:         handler.NewArtHandler(artSvc, profileSvc),
		Event:       handler.NewEventHandler(eventsSvc),
		Onboarding:  handler.NewOnboardingHandler(onboardingSvc),
		Auth:        handler.NewAuthHandler(authSvc, cfg.AuthCookieName),
	}

	router := httpadapter.NewRouter(cfg, handlers, httpadapter.RouterDeps{
		Identity: identitySvc,
		Auth:     authCfg,
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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
}
