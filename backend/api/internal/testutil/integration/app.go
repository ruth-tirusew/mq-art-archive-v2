//go:build integration

package integration

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mq/api/config"
	"github.com/mq/api/internal/adapter/driven/auth"
	eventsadapter "github.com/mq/api/internal/adapter/driven/events"
	"github.com/mq/api/internal/adapter/driven/persistence/postgres"
	httpadapter "github.com/mq/api/internal/adapter/driving/http"
	"github.com/mq/api/internal/adapter/driving/http/handler"
	"github.com/mq/api/internal/adapter/driving/http/middleware"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/port/inbound"
	artuc "github.com/mq/api/internal/usecase/art"
	authuc "github.com/mq/api/internal/usecase/auth"
	contentuc "github.com/mq/api/internal/usecase/content"
	eventsuc "github.com/mq/api/internal/usecase/events"
	identityuc "github.com/mq/api/internal/usecase/identity"
	institutionuc "github.com/mq/api/internal/usecase/institution"
	onboardinguc "github.com/mq/api/internal/usecase/onboarding"
	profileuc "github.com/mq/api/internal/usecase/profile"
)

// App wires real Postgres repositories to use-case services, mirroring cmd/api.
type App struct {
	Pool        *postgres.Pool
	Content     inbound.ContentService
	Art         inbound.ArtService
	Profile     inbound.ProfileService
	Identity    inbound.IdentityService
	Institution inbound.InstitutionService
	Onboarding  inbound.OnboardingService
	Events      inbound.EventsService
	Auth        inbound.AuthService
	TokenSvc    *auth.TokenService
}

func NewApp(t *testing.T) *App {
	t.Helper()

	pool := SetupPostgresPool(t)

	articleRepo := postgres.NewArticleRepository(pool)
	artPostRepo := postgres.NewArtPostRepository(pool)
	profileRepo := postgres.NewProfileRepository(pool)
	userRepo := postgres.NewUserRepository(pool)
	oauthAccountRepo := postgres.NewOAuthAccountRepository(pool)
	institutionRepo := postgres.NewInstitutionRepository(pool)
	onboardingRepo := postgres.NewOnboardingRepository(pool)
	eventRepo := postgres.NewEventRepository(pool)
	eventLocationRepo := postgres.NewEventLocationRepository(pool)
	eventSource := eventsadapter.NewScraperNoop()

	tokenSvc := auth.NewTokenService("integration-test-secret", config.Load().JWTAccessTTL)
	stateSvc := auth.NewOAuthStateService("integration-test-secret", config.Load().JWTAccessTTL)
	identitySvc := identityuc.NewService(userRepo)
	authSvc := authuc.NewService(userRepo, oauthAccountRepo, tokenSvc, stateSvc, nil, []string{"http://localhost:5173"})

	return &App{
		Pool:        pool,
		Content:     contentuc.NewService(articleRepo),
		Art:         artuc.NewService(artPostRepo),
		Profile:     profileuc.NewService(profileRepo),
		Identity:    identitySvc,
		Institution: institutionuc.NewService(institutionRepo),
		Onboarding:  onboardinguc.NewService(onboardingRepo),
		Events:      eventsuc.NewService(eventRepo, eventLocationRepo, eventSource),
		Auth:        authSvc,
		TokenSvc:    tokenSvc,
	}
}

func (a *App) Router() *gin.Engine {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		CORSOrigins:    []string{"http://localhost:5173"},
		AuthCookieName: "mq_access_token",
		AuthDevMode:    true,
	}
	authCfg := middleware.AuthConfig{
		Verifier:   a.TokenSvc,
		Identity:   a.Identity,
		CookieName: cfg.AuthCookieName,
		DevMode:    cfg.AuthDevMode,
	}

	handlers := httpadapter.Handlers{
		Health:      handler.NewHealthHandler(),
		Article:     handler.NewArticleHandler(a.Content),
		Profile:     handler.NewProfileHandler(a.Profile),
		Institution: handler.NewInstitutionHandler(a.Institution),
		Art:         handler.NewArtHandler(a.Art, a.Profile),
		Onboarding:  handler.NewOnboardingHandler(a.Onboarding),
		Auth:        handler.NewAuthHandler(a.Auth, cfg.AuthCookieName),
	}
	return httpadapter.NewRouter(cfg, handlers, httpadapter.RouterDeps{
		Identity: a.Identity,
		Auth:     authCfg,
	})
}

func InsertAdminUser(t *testing.T, pool *postgres.Pool) uuid.UUID {
	t.Helper()
	id := uuid.New()
	InsertUser(t, pool, id, identity.RoleAdmin)
	return id
}
