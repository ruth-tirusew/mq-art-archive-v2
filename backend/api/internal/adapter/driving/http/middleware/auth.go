package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mq/api/internal/adapter/driving/http/requestauth"
	"github.com/mq/api/internal/domain/apperrors"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/port/inbound"
	"github.com/mq/api/internal/port/outbound"
)

type AuthConfig struct {
	Verifier   outbound.TokenVerifier
	Identity   inbound.IdentityService
	CookieName string
	DevMode    bool
}

func Authenticate(cfg AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if user, ok := authenticate(cfg, c); ok {
			setUserContext(c, user)
			c.Next()
			return
		}

		if cfg.DevMode {
			if id, err := requestauth.UserIDFromHeader(c.GetHeader("X-User-ID")); err == nil {
				user, err := cfg.Identity.GetUser(c.Request.Context(), id)
				if err == nil {
					setUserContext(c, user)
					c.Next()
					return
				}
			}
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	}
}

func authenticate(cfg AuthConfig, c *gin.Context) (*identity.User, bool) {
	token := tokenFromRequest(c, cfg.CookieName)
	if token == "" {
		return nil, false
	}

	claims, err := cfg.Verifier.Verify(c.Request.Context(), token)
	if err != nil {
		return nil, false
	}

	user, err := cfg.Identity.GetUser(c.Request.Context(), claims.UserID)
	if err != nil {
		return nil, false
	}
	return user, true
}

func tokenFromRequest(c *gin.Context, cookieName string) string {
	auth := c.GetHeader("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	if cookie, err := c.Cookie(cookieName); err == nil && cookie != "" {
		return cookie
	}
	return ""
}

func setUserContext(c *gin.Context, user *identity.User) {
	c.Set(requestauth.ContextUserID, user.ID)
	c.Set(requestauth.ContextUserRole, user.Role)
	c.Set(requestauth.ContextUserEmail, user.Email)
}

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, ok := c.Get(requestauth.ContextUserRole)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if string(userRole.(identity.Role)) != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": apperrors.ErrForbidden.Error()})
			return
		}
		c.Next()
	}
}

func RequireAnyRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}
	return func(c *gin.Context) {
		value, ok := c.Get(requestauth.ContextUserRole)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if _, ok := allowed[string(value.(identity.Role))]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": apperrors.ErrForbidden.Error()})
			return
		}
		c.Next()
	}
}

// OptionalAuthenticate sets user context when a valid token is present but does not reject.
func OptionalAuthenticate(cfg AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if user, ok := authenticate(cfg, c); ok {
			setUserContext(c, user)
		} else if cfg.DevMode {
			if id, err := requestauth.UserIDFromHeader(c.GetHeader("X-User-ID")); err == nil {
				if user, err := cfg.Identity.GetUser(c.Request.Context(), id); err == nil {
					setUserContext(c, user)
				}
			}
		}
		c.Next()
	}
}
