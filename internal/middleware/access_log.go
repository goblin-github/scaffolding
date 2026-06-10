package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"scaffolding/pkg/logger"
)

func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		elapsed := time.Since(start)
		// X-Process-Time：毫秒数，方便客户端做性能监控
		c.Header("X-Process-Time", strconv.FormatInt(elapsed.Milliseconds(), 10)+"ms")

		logger.Info(c.Request.Context(), "http_access",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency", elapsed.String(),
			"client_ip", c.ClientIP(),
		)
	}
}
