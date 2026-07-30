package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mq/api/internal/port/outbound"
)

func Recovery(monitor outbound.ErrorMonitor) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				err := fmt.Errorf("panic: %v", recovered)
				requestID, _ := c.Get("request_id")
				if monitor != nil {
					monitor.Capture(err, map[string]string{"request_id": fmt.Sprint(requestID)})
				}
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			}
		}()
		c.Next()
	}
}
