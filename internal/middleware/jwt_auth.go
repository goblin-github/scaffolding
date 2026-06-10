package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"scaffolding/pkg/auth"
	"scaffolding/pkg/cache"

	"scaffolding/internal/errcode"
	"scaffolding/pkg/response"
)

// JWTAuth 返回 JWT 鉴权中间件。
// cache 用于 Layer 2 JTI 黑名单检查。如果不需要黑名单功能，传 nil 即可。
func JWTAuth(jwtMgr *auth.JWTManager, cache *cache.Cache) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			response.Error(c, errcode.ErrUnauthorized)
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")

		// JTI 黑名单检查：查询 Redis 是否存在 "jti:<jti>" key
		var jtiBlacklist func(jti string) bool
		if cache != nil {
			jtiBlacklist = func(jti string) bool {
				exists, _ := cache.Exists(c.Request.Context(), "jti:"+jti)
				return exists > 0
			}
		}

		claims, err := jwtMgr.ValidateToken(tokenStr, jtiBlacklist)
		if err != nil {
			response.Error(c, errcode.ErrUnauthorized)
			c.Abort()
			return
		}

		// 把用户信息注入 context
		c.Set("user_id", claims.UserID)
		c.Set("token_version", claims.TokenVersion)
		c.Next()
	}
}

// GetCurrentUserID 从 gin.Context 中提取当前登录用户 ID。
func GetCurrentUserID(c *gin.Context) uint64 {
	id, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	uid, _ := id.(uint64)
	return uid
}
