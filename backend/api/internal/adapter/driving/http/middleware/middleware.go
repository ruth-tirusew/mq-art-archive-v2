package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = uuid.New().String()
		}
		c.Header("X-Request-ID", id)
		c.Set("request_id", id)
		c.Next()
	}
}

// CORS allows the configured origins. devMode adds the X-User-ID header, which only exists for
// local/test header authentication and must never be advertised by a production API.
func CORS(origins []string, devMode bool) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(origins))
	for _, o := range origins {
		allowed[strings.TrimSpace(o)] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if _, ok := allowed[origin]; ok {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			allowHeaders := "Content-Type, Authorization, X-Request-ID"
			if devMode {
				allowHeaders = "Content-Type, Authorization, X-User-ID, X-Request-ID"
			}
			c.Header("Access-Control-Allow-Headers", allowHeaders)
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
