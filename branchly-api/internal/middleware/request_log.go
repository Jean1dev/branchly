package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		slog.Info("request",
			"method", c.Request.Method,
			"path", path,
			slog.Duration("duration", time.Since(start)),
		)
	}
}
