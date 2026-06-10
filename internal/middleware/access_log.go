package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"scaffolding/pkg/logger"
)

func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		elapsed := time.Since(start)
		logger.Info(c.Request.Context(), "ACCESS",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency", elapsed.String(),
			"client_ip", c.ClientIP(),
		)
	}
}
