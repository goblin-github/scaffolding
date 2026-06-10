package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"scaffolding/pkg/logger"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-ID")
		if rid == "" {
			rid = uuid.New().String()
		}
		c.Request = c.Request.WithContext(
			context.WithValue(c.Request.Context(), logger.TraceIDKey, rid),
		)
		c.Set("trace_id", rid) // 供 response 包直接取值
		c.Header("X-Request-ID", rid)
		c.Next()
	}
}
