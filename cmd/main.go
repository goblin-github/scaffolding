package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"scaffolding/internal"
	"scaffolding/internal/config"
	"scaffolding/pkg/logger"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load("configs/config.yml")
	if err != nil {
		println("CRITICAL: Failed to load configuration:", err.Error())
		os.Exit(1)
	}

	// 2. 初始化日志（必须在 InitAPI 之前，InitAPI 内部用 slog）
	logger.InitLogger(logger.Options{
		Filename:   cfg.Logger.Filename,
		Level:      cfg.Logger.Level,
		MaxSize:    cfg.Logger.MaxSize,
		MaxBackups: cfg.Logger.MaxBackups,
		MaxAge:     cfg.Logger.MaxAge,
		Compress:   cfg.Logger.Compress,
		Console:    cfg.Logger.Console,
	})
	logger.Info(nil, "Logger initialized")

	// 3. 初始化基础设施依赖（DB、Redis、JWT、i18n）→ 组装 Registry
	reg, cleanup, err := internal.InitRegistry(cfg)
	if err != nil {
		logger.Error(nil, "Infra init failed", "err", err)
		os.Exit(1)
	}
	defer cleanup()

	// 4. 装配调用链 → HTTP Server
	srv := internal.NewServer(reg)

	// 5. 启动 + 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info(nil, "Server starting", "addr", cfg.App.Port)
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error(nil, "Server crashed", "err", err)
			os.Exit(1)
		}
	}()

	sig := <-quit
	logger.Info(nil, "Shutting down server", "signal", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.App.ShutdownTimeout)*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error(nil, "Server forced to shutdown", "err", err)
		os.Exit(1)
	}

	if err := logger.Sync(); err != nil {
		println("WARN: logger sync failed:", err.Error())
	}
	logger.Info(nil, "Server exited gracefully")
}
