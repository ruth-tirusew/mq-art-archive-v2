package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders hardens every response. This API only serves JSON, so it can be strict: nothing
// may frame it, content types are not sniffed, and no resource may load from it.
// hsts should be true when the API is served over HTTPS (production), where it pins browsers to it.
func SecurityHeaders(hsts bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		if hsts {
			h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		c.Next()
	}
}

// BodyLimit caps request bodies. Unbounded bodies let one client tie up memory (JSON binding reads
// the whole thing); handlers see a normal bind error when the cap is exceeded.
func BodyLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
	}
}
