package http

import (
	"github.com/gin-gonic/gin"
	"github.com/mq/api/config"
	"github.com/mq/api/internal/adapter/driving/http/handler"
	"github.com/mq/api/internal/adapter/driving/http/middleware"
	"github.com/mq/api/internal/port/inbound"
)

type Handlers struct {
	Health      *handler.HealthHandler
	Article     *handler.ArticleHandler
	Profile     *handler.ProfileHandler
	Institution *handler.InstitutionHandler
	Art         *handler.ArtHandler
	Event       *handler.EventHandler
	Onboarding  *handler.OnboardingHandler
	Auth        *handler.AuthHandler
}

type RouterDeps struct {
	Identity inbound.IdentityService
	Auth     middleware.AuthConfig
}

func NewRouter(cfg config.Config, handlers Handlers, deps RouterDeps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), middleware.RequestID(), middleware.CORS(cfg.CORSOrigins))

	r.GET("/health", handlers.Health.Health)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/articles", handlers.Article.List)
		v1.GET("/articles/:slug", handlers.Article.GetBySlug)
		v1.POST("/articles", middleware.Authenticate(deps.Auth), handlers.Article.Create)

		v1.GET("/artists", handlers.Profile.List)
		v1.GET("/artists/:slug", handlers.Profile.GetBySlug)
		v1.GET("/artists/:slug/posts", handlers.Art.ListByArtistSlug)
		v1.GET("/profiles/@"+":handle", handlers.Profile.GetByHandle)

		v1.GET("/institutions/:slug", handlers.Institution.GetBySlug)
		v1.GET("/posts", handlers.Art.List)
		v1.GET("/posts/:id", handlers.Art.GetByID)
		v1.POST("/posts", middleware.Authenticate(deps.Auth), handlers.Art.Create)

		v1.GET("/events", handlers.Event.List)
		v1.GET("/events/:slug", handlers.Event.GetBySlug)

		auth := v1.Group("/auth")
		{
			auth.GET("/google", handlers.Auth.GoogleLogin)
			auth.GET("/google/callback", handlers.Auth.GoogleCallback)
			auth.POST("/logout", handlers.Auth.Logout)
			auth.GET("/me", middleware.OptionalAuthenticate(deps.Auth), handlers.Auth.Me)
		}
	}

	admin := r.Group("/admin/v1")
	admin.Use(middleware.Authenticate(deps.Auth), middleware.RequireRole("admin"))
	{
		admin.GET("/applications", handlers.Onboarding.ListPending)
		admin.GET("/applications/:id", handlers.Onboarding.GetByID)
		admin.PUT("/applications/:id", handlers.Onboarding.Review)
	}

	return r
}
