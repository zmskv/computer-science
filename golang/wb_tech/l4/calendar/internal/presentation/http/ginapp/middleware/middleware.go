package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"calendar/internal/domain/interfaces"
)

func LoggerMiddleware(logger interfaces.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		method := c.Request.Method
		path := c.Request.URL.Path

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()
		logger.Info("request completed", map[string]any{
			"method":   method,
			"path":     path,
			"status":   status,
			"duration": duration.String(),
		})
	}
}
