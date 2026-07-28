package http

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mq/api/config"
	"github.com/mq/api/internal/adapter/driving/http/handler"
	"github.com/mq/api/internal/adapter/driving/http/middleware"
	"github.com/mq/api/internal/port/inbound"
	"github.com/mq/api/internal/port/outbound"
)

type Handlers struct {
	Health      *handler.HealthHandler
	Article     *handler.ArticleHandler
	Profile     *handler.ProfileHandler
	Institution *handler.InstitutionHandler
	Art         *handler.ArtHandler
	Event       *handler.EventHandler
	Search      *handler.SearchHandler
	Onboarding  *handler.OnboardingHandler
	Auth        *handler.AuthHandler
	Settings    *handler.SettingsHandler
	Media       *handler.MediaHandler
	Wiki        *handler.WikiHandler
	Analytics   *handler.AnalyticsHandler
	UserAdmin   *handler.UserAdminHandler
}

type RouterDeps struct {
	Identity inbound.IdentityService
	Auth     middleware.AuthConfig
	Monitor  outbound.ErrorMonitor
}

func NewRouter(cfg config.Config, handlers Handlers, deps RouterDeps) *gin.Engine {
	r := gin.New()
	r.Use(middleware.RequestID(), middleware.Recovery(deps.Monitor), middleware.CORS(cfg.CORSOrigins))
	authLimit := middleware.NewRateLimiter(10, time.Minute).Middleware()
	writeLimit := middleware.NewRateLimiter(30, time.Minute).Middleware()

	r.GET("/health", handlers.Health.Health)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/articles", handlers.Article.List)
		v1.GET("/articles/:slug", handlers.Article.GetBySlug)
		v1.POST("/articles", middleware.Authenticate(deps.Auth), middleware.RequireRole("admin"), handlers.Article.Create)

		v1.GET("/artists", handlers.Profile.List)
		v1.GET("/artists/:slug", handlers.Profile.GetBySlug)
		v1.GET("/artists/:slug/posts", handlers.Art.ListByArtistSlug)
		v1.GET("/profiles/@"+":handle", handlers.Profile.GetByHandle)

		v1.GET("/institutions/:slug", handlers.Institution.GetBySlug)
		v1.GET("/posts", handlers.Art.List)
		v1.GET("/posts/:id", handlers.Art.GetByID)
		v1.POST("/posts", middleware.Authenticate(deps.Auth), handlers.Art.Create)

		v1.GET("/events", handlers.Event.List)
		v1.POST("/events/submissions", writeLimit, handlers.Event.Submit)
		v1.GET("/events/:slug", handlers.Event.GetBySlug)
		v1.GET("/search", handlers.Search.Search)
		v1.GET("/handles/:handle/available", handlers.Onboarding.CheckHandleAvailable)
		v1.POST("/analytics/view", writeLimit, handlers.Analytics.Record)

		v1.POST("/applications", writeLimit, middleware.Authenticate(deps.Auth), handlers.Onboarding.Submit)
		v1.GET("/applications/me", middleware.Authenticate(deps.Auth), handlers.Onboarding.GetMyApplication)

		me := v1.Group("/me")
		me.Use(middleware.Authenticate(deps.Auth), middleware.RequireRole("artist"))
		{
			me.GET("/profile", handlers.Profile.GetMyProfile)
			me.PUT("/profile", handlers.Profile.UpdateMyProfile)
			me.GET("/posts", handlers.Art.ListMyPosts)
			me.GET("/posts/:id", handlers.Art.GetMyPost)
			me.PUT("/posts/:id", handlers.Art.UpdateMyPost)
			me.POST("/posts/:id/publish", handlers.Art.PublishMyPost)
			me.POST("/posts/:id/unpublish", handlers.Art.UnpublishMyPost)
			me.POST("/posts/:id/archive", handlers.Art.ArchiveMyPost)
			me.DELETE("/posts/:id", handlers.Art.DeleteMyPost)
			me.POST("/media/sign", writeLimit, handlers.Media.Sign)
			me.POST("/media/complete", writeLimit, handlers.Media.Complete)
			me.GET("/analytics", handlers.Analytics.MeOverview)
		}

		wiki := v1.Group("/me/wiki")
		wiki.Use(middleware.Authenticate(deps.Auth), middleware.RequireAnyRole("artist", "contributor"))
		wiki.POST("/submissions", writeLimit, handlers.Wiki.Submit)
		wiki.GET("/submissions", handlers.Wiki.ListMine)

		auth := v1.Group("/auth")
		{
			auth.GET("/google", handlers.Auth.GoogleLogin)
			auth.GET("/google/callback", handlers.Auth.GoogleCallback)
			auth.POST("/register", handlers.Auth.Register)
			auth.POST("/login", authLimit, handlers.Auth.Login)
			auth.POST("/logout", handlers.Auth.Logout)
			auth.POST("/forgot-password", authLimit, handlers.Auth.ForgotPassword)
			auth.POST("/reset-password", authLimit, handlers.Auth.ResetPassword)
			auth.POST("/verify-email", authLimit, handlers.Auth.VerifyEmail)
			auth.POST("/resend-verification", authLimit, middleware.Authenticate(deps.Auth), handlers.Auth.ResendEmailVerification)
			auth.GET("/me", middleware.Authenticate(deps.Auth), handlers.Auth.Me)
			auth.PUT("/me/profile", middleware.Authenticate(deps.Auth), handlers.Auth.UpdateMyProfile)
			auth.PUT("/me/email", middleware.Authenticate(deps.Auth), handlers.Auth.ChangeEmail)
			auth.PUT("/me/password", middleware.Authenticate(deps.Auth), handlers.Auth.ChangePassword)
			auth.GET("/me/notifications", middleware.Authenticate(deps.Auth), handlers.Auth.GetNotifications)
			auth.PUT("/me/notifications", middleware.Authenticate(deps.Auth), handlers.Auth.UpdateNotifications)
		}
	}

	admin := r.Group("/admin/v1")
	admin.Use(middleware.Authenticate(deps.Auth), middleware.RequireRole("admin"))
	{
		admin.GET("/applications", handlers.Onboarding.ListPending)
		admin.GET("/applications/:id", handlers.Onboarding.GetByID)
		admin.PUT("/applications/:id", handlers.Onboarding.Review)
		admin.POST("/media/sign", writeLimit, handlers.Media.Sign)
		admin.POST("/media/complete", writeLimit, handlers.Media.Complete)
		admin.GET("/wiki/submissions", handlers.Wiki.ListPending)
		admin.POST("/wiki/submissions/:id/approve", handlers.Wiki.Approve)
		admin.POST("/wiki/submissions/:id/reject", handlers.Wiki.Reject)
		admin.GET("/analytics", handlers.Analytics.Query)
		admin.GET("/users", handlers.UserAdmin.List)
		admin.PATCH("/users/:id", handlers.UserAdmin.UpdateRole)

		admin.GET("/artists", handlers.Profile.ListArtistsAdmin)
		admin.POST("/artists", handlers.Profile.CreateArtistAdmin)
		admin.GET("/artists/:id", handlers.Profile.GetArtistAdmin)
		admin.PUT("/artists/:id", handlers.Profile.UpdateArtistAdmin)
		admin.PATCH("/artists/:id", handlers.Profile.PatchArtistAdmin)
		admin.DELETE("/artists/:id", handlers.Profile.DeleteArtistAdmin)

		admin.GET("/posts", handlers.Art.ListPostsAdmin)
		admin.POST("/posts", handlers.Art.CreatePostAdmin)
		admin.GET("/posts/:id", handlers.Art.GetPostAdmin)
		admin.PUT("/posts/:id", handlers.Art.UpdatePostAdmin)
		admin.PATCH("/posts/:id", handlers.Art.PatchPostAdmin)
		admin.DELETE("/posts/:id", handlers.Art.DeletePostAdmin)

		admin.GET("/articles", handlers.Article.ListArticlesAdmin)
		admin.GET("/articles/:id/revisions", handlers.Article.ListArticleRevisionsAdmin)
		admin.GET("/articles/:id/revisions/:version", handlers.Article.GetArticleRevisionAdmin)
		admin.POST("/articles/:id/revisions/:version/restore", handlers.Article.RestoreArticleRevisionAdmin)
		admin.GET("/articles/:id", handlers.Article.GetArticleAdmin)
		admin.POST("/articles", handlers.Article.CreateArticleAdmin)
		admin.PUT("/articles/:id", handlers.Article.UpdateArticleAdmin)
		admin.PATCH("/articles/:id", handlers.Article.PatchArticleAdmin)

		admin.GET("/events", handlers.Event.ListAdmin)
		admin.POST("/events", handlers.Event.CreateAdmin)
		admin.GET("/events/:id", handlers.Event.GetByID)
		admin.PUT("/events/:id", handlers.Event.UpdateAdmin)
		admin.PATCH("/events/:id", handlers.Event.Review)
		admin.DELETE("/events/:id", handlers.Event.DeleteAdmin)
		admin.POST("/events/sync", handlers.Event.Sync)

		admin.GET("/settings/scrape", handlers.Settings.GetScrape)
		admin.PUT("/settings/scrape", handlers.Settings.UpdateScrape)
	}

	return r
}
