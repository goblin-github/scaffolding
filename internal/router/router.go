package router

import (
	"github.com/gin-gonic/gin"

	"scaffolding/internal/handler"
	"scaffolding/internal/middleware"
)

// Dependencies 聚合所有 HTTP handler 和中间件，由上层（internal/server.go）装配后传入。
type Dependencies struct {
	Article *handler.ArticleHandler
	JWTAuth gin.HandlerFunc
	// 后续加新的 handler 在这里继续加字段
}

func NewEngine(mode string, deps Dependencies) *gin.Engine {
	gin.SetMode(mode)
	engine := gin.New()

	// 全局中间件
	engine.Use(middleware.RequestID())
	engine.Use(middleware.CORS())
	engine.Use(middleware.AccessLog())
	engine.Use(gin.Recovery())

	registerRoutes(engine, deps)
	return engine
}

func registerRoutes(r *gin.Engine, deps Dependencies) {
	// 健康检查，不需要认证

	v1 := r.Group("/api/v1")
	registerArticleRoutes(v1, deps)
}

// registerArticleRoutes 示例：文章模块路由。
// 每个业务模块一个私有函数，模块内自己决定哪些接口公开、哪些需要认证。
func registerArticleRoutes(r *gin.RouterGroup, deps Dependencies) {
	articles := r.Group("/articles")

	// 需要认证的接口
	if deps.JWTAuth != nil {
		articles.Use(deps.JWTAuth)
	}

	articles.POST("", deps.Article.Create)
	articles.GET("/:id", deps.Article.Get)
}
