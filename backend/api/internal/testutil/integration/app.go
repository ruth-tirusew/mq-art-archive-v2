//go:build integration

package integration

import (
	"context"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mq/api/config"
	"github.com/mq/api/internal/adapter/driven/auth"
	eventsadapter "github.com/mq/api/internal/adapter/driven/events"
	maileradapter "github.com/mq/api/internal/adapter/driven/mailer"
	"github.com/mq/api/internal/adapter/driven/persistence/postgres"
	provisioningadapter "github.com/mq/api/internal/adapter/driven/provisioning"
	"github.com/mq/api/internal/adapter/driven/richtext"
	httpadapter "github.com/mq/api/internal/adapter/driving/http"
	"github.com/mq/api/internal/adapter/driving/http/handler"
	"github.com/mq/api/internal/adapter/driving/http/middleware"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/port/inbound"
	artuc "github.com/mq/api/internal/usecase/art"
	authuc "github.com/mq/api/internal/usecase/auth"
	contentuc "github.com/mq/api/internal/usecase/content"
	digestuc "github.com/mq/api/internal/usecase/digest"
	eventsuc "github.com/mq/api/internal/usecase/events"
	identityuc "github.com/mq/api/internal/usecase/identity"
	institutionuc "github.com/mq/api/internal/usecase/institution"
	onboardinguc "github.com/mq/api/internal/usecase/onboarding"
	profileuc "github.com/mq/api/internal/usecase/profile"
	questionnaireuc "github.com/mq/api/internal/usecase/questionnaire"
	searchuc "github.com/mq/api/internal/usecase/search"
	settingsuc "github.com/mq/api/internal/usecase/settings"
)

// App wires real Postgres repositories to use-case services, mirroring cmd/api.
type App struct {
	Pool          *postgres.Pool
	Content       inbound.ContentService
	Art           inbound.ArtService
	Profile       inbound.ProfileService
	Identity      inbound.IdentityService
	Institution   inbound.InstitutionService
	Onboarding    *onboardinguc.Service
	Questionnaire *questionnaireuc.Service
	Mailer        *RecordingMailer
	Events        inbound.EventsService
	Search        inbound.SearchService
	Auth          inbound.AuthService
	Settings      inbound.SettingsService
	Digest        inbound.DigestService
	TokenSvc      *auth.TokenService
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
	passwordResetRepo := postgres.NewPasswordResetRepository(pool)
	notifPrefsRepo := postgres.NewNotificationPreferencesRepository(pool)
	scrapeSettingsRepo := postgres.NewScrapeSettingsRepository(pool)
	eventRepo := postgres.NewEventRepository(pool)
	eventLocationRepo := postgres.NewEventLocationRepository(pool)
	eventSource := eventsadapter.NewScraperNoop()
	digestRecipientRepo := postgres.NewDigestRecipientRepository(pool)
	digestRunRepo := postgres.NewDigestRunRepository(pool)
	telegramLinkRepo := postgres.NewTelegramLinkRepository(pool)

	tokenSvc := auth.NewTokenService("integration-test-secret", config.Load().JWTAccessTTL)
	stateSvc := auth.NewOAuthStateService("integration-test-secret", config.Load().JWTAccessTTL)
	identitySvc := identityuc.NewService(userRepo)
	settingsSvc := settingsuc.NewService(scrapeSettingsRepo, nil, "")
	authSvc := authuc.NewService(
		userRepo,
		oauthAccountRepo,
		tokenSvc,
		auth.NewBcryptPasswordHasher(),
		stateSvc,
		nil,
		[]string{"http://localhost:5173"},
		nil,
		passwordResetRepo,
		notifPrefsRepo,
		"http://localhost:5173",
	)

	mailer := &RecordingMailer{}
	questionnaireSvc := questionnaireuc.NewService(
		postgres.NewQuestionnaireRepository(pool),
		postgres.NewQuestionnaireAnswerRepository(pool),
		postgres.NewQuestionnaireNudgeRepository(pool),
		richtext.NewSanitizer(),
		userRepo,
		questionnaireuc.WithMailer(mailer, "http://localhost:5173"),
		questionnaireuc.WithLookups(profileRepo, onboardingRepo),
	)

	return &App{
		Pool:          pool,
		Questionnaire: questionnaireSvc,
		Mailer:        mailer,
		Content:       contentuc.NewService(articleRepo),
		Art:           artuc.NewService(artPostRepo, artuc.WithRichText(richtext.NewSanitizer())),
		Profile:       profileuc.NewService(profileRepo, userRepo, artPostRepo, profileuc.WithRichText(richtext.NewSanitizer())),
		Identity:      identitySvc,
		Institution:   institutionuc.NewService(institutionRepo),
		Onboarding: onboardinguc.NewService(onboardingRepo, userRepo, profileRepo, institutionRepo,
			onboardinguc.WithMailer(mailer, "http://localhost:5173"),
			onboardinguc.WithInvites(postgres.NewInviteRepository(pool)),
			onboardinguc.WithQuestionnaire(questionnaireSvc),
			onboardinguc.WithInvitedRegistrar(authSvc),
			onboardinguc.WithApplicantWorks(postgres.NewApplicantWorkRepository(pool)),
			onboardinguc.WithArtistProvisioner(provisioningadapter.NewArtistProvisioner(profileRepo, artPostRepo)),
		),
		Events:   eventsuc.NewService(eventRepo, eventLocationRepo, eventSource),
		Search:   searchuc.NewService(articleRepo, eventRepo),
		Auth:     authSvc,
		Settings: settingsSvc,
		Digest: digestuc.NewService(
			articleRepo, eventRepo, artPostRepo, digestRecipientRepo, digestRunRepo,
			notifPrefsRepo, telegramLinkRepo, "integration-test-unsubscribe-secret",
			maileradapter.NewLogMailer(true), nil, "http://localhost:5173", "http://localhost:8080",
		),
		TokenSvc: tokenSvc,
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
		Health:        handler.NewHealthHandler(),
		Article:       handler.NewArticleHandler(a.Content),
		Profile:       handler.NewProfileHandler(a.Profile),
		Institution:   handler.NewInstitutionHandler(a.Institution),
		Art:           handler.NewArtHandler(a.Art, a.Profile),
		Event:         handler.NewEventHandler(a.Events),
		Search:        handler.NewSearchHandler(a.Search),
		Onboarding:    handler.NewOnboardingHandler(a.Onboarding),
		Questionnaire: handler.NewQuestionnaireHandler(a.Questionnaire, a.Questionnaire),
		Invite:        handler.NewInviteHandler(a.Onboarding, cfg.AuthCookieName, "http://localhost:5173"),
		Works:         handler.NewApplicantWorksHandler(a.Onboarding),
		Auth:          handler.NewAuthHandler(a.Auth, cfg.AuthCookieName),
		Settings:      handler.NewSettingsHandler(a.Settings),
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

// RecordingMailer captures outgoing mail so tests can assert on it without sending anything.
type RecordingMailer struct {
	mu   sync.Mutex
	Sent []SentMail
}

type SentMail struct{ To, Subject, Body string }

func (m *RecordingMailer) Send(ctx context.Context, to, subject, body string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Sent = append(m.Sent, SentMail{To: to, Subject: subject, Body: body})
	return nil
}

func (m *RecordingMailer) SendHTML(ctx context.Context, to, subject, htmlBody, textBody string) error {
	return m.Send(ctx, to, subject, textBody)
}

// Messages returns a copy of what has been sent so far.
func (m *RecordingMailer) Messages() []SentMail {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]SentMail(nil), m.Sent...)
}
