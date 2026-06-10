package internal

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"scaffolding/internal/handler"
	"scaffolding/internal/middleware"
	"scaffolding/internal/repository"
	"scaffolding/internal/router"
	"scaffolding/internal/service"
)

// Server 封装 HTTP 服务。
type Server struct {
	httpServer *http.Server
}

// NewServer 从 Registry 装配完整的调用链并返回可启动的 Server。
// main 调用流程：config.Load → logger.InitLogger → config.InitAPI → internal.NewServer。
func NewServer(reg *Registry) *Server {
	// ── 装配调用链：repo → service → handler ──

	articleRepo := repository.NewArticleRepository(reg.DB)
	articleSvc := service.NewArticleService(articleRepo)
	articleHandler := handler.NewArticleHandler(articleSvc)

	// ── 装配认证中间件（JWT 为空时跳过）──

	var authMw gin.HandlerFunc
	if reg.JWT != nil {
		authMw = middleware.JWTAuth(reg.JWT, reg.Cache)
	}

	// ── 装配路由 ──
	deps := router.Dependencies{
		Article: articleHandler,
		JWTAuth: authMw,
	}
	engine := router.NewEngine(reg.Config.App.Mode, deps)

	appCfg := reg.Config.App
	return &Server{
		httpServer: &http.Server{
			Addr:         appCfg.Port,
			Handler:      engine,
			ReadTimeout:  time.Duration(appCfg.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(appCfg.WriteTimeout) * time.Second,
			IdleTimeout:  time.Duration(appCfg.IdleTimeout) * time.Second,
		},
	}
}

// Start 启动 HTTP 服务，阻塞直到 Shutdown 调用。
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// Shutdown 优雅关闭，等待正在处理的请求完成。
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
