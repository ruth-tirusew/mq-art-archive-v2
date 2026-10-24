package middleware

import (
	"mime"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CSRF blocks cross-site request forgery against cookie-authenticated, state-changing routes.
//
// The session cookie is SameSite=None in production (the web apps and the API sit on different
// sites), so the browser attaches it to requests made from any page. Two checks stop a hostile
// page from using that:
//
//  1. A browser always sends Origin on cross-origin POST/PUT/PATCH/DELETE. If it is present it must
//     be one of the configured origins. Requests with no Origin (curl, server-to-server, tests)
//     are not browser-driven cross-site requests, so they are not CSRF and pass through.
//  2. A request with a body must declare application/json. Otherwise a page can send a "simple"
//     cross-origin request (text/plain, form-encoded) that skips the CORS preflight entirely and
//     still has JSON the handler will happily parse.
//
// Safe methods (GET, HEAD, OPTIONS) are never blocked; OPTIONS is the CORS preflight.
func CSRF(origins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(origins))
	for _, o := range origins {
		allowed[strings.TrimSpace(o)] = struct{}{}
	}

	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			c.Next()
			return
		}

		if origin := c.GetHeader("Origin"); origin != "" {
			if _, ok := allowed[origin]; !ok {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "cross-site request blocked"})
				return
			}
		}

		if hasBody(c.Request) && !isJSON(c.GetHeader("Content-Type")) {
			c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, gin.H{"error": "content type must be application/json"})
			return
		}
		c.Next()
	}
}

func hasBody(r *http.Request) bool {
	return r.ContentLength > 0 || len(r.TransferEncoding) > 0
}

func isJSON(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	return err == nil && mediaType == "application/json"
}
