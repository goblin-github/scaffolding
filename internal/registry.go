package internal

import (
	"fmt"
	"os"
	"scaffolding/internal/config"
	"scaffolding/pkg/database"
	"scaffolding/pkg/logger"
	"time"

	"gorm.io/gorm"

	"scaffolding/pkg/auth"
	"scaffolding/pkg/cache"
)

// Registry 聚合所有基础设施依赖，由 InitAPI 组装。
// internal/server.go 从 Registry 装配调用链。
type Registry struct {
	Config *config.Config
	DB     *gorm.DB
	Cache  *cache.Cache
	JWT    *auth.JWTManager
	I18n   *config.Bundle
}

func InitRegistry(cfg *config.Config) (*Registry, func(), error) {
	// ── 数据库 ──
	db, err := database.NewMySQL(database.Options{
		DSN:             cfg.Database.DSN,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("database: %w", err)
	}

	// ── Redis ──
	c, err := cache.NewRedis(cache.Options{
		Addr:     cfg.Cache.Addr,
		Password: cfg.Cache.Password,
		DB:       cfg.Cache.DB,
		PoolSize: cfg.Cache.PoolSize,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("redis: %w", err)
	}

	// ── i18n ──
	bundle := config.InitBundle()
	loadLocales(bundle)

	// ── JWT（secret 为空则跳过，认证中间件不生效）──
	var jwtMgr *auth.JWTManager
	if cfg.JWT.Secret != "" {
		jwtMgr = auth.NewJWTManager(
			cfg.JWT.Secret,
			time.Duration(cfg.JWT.ExpiresAfter)*time.Second,
			nil, // TokenVersionStore 可选，暂时不注入
		)
	}

	reg := &Registry{
		Config: cfg,
		DB:     db,
		Cache:  c,
		JWT:    jwtMgr,
		I18n:   bundle,
	}

	// cleanup 按初始化的逆序释放资源
	cleanup := func() {
		if err := c.Close(); err != nil {
			logger.Error(nil, "close Redis failed", "err", err)
		}
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			err := sqlDB.Close()
			if err != nil {
				logger.Error(nil, "close Database failed: %w", err)
			}
		}
	}
	return reg, cleanup, nil
}

// loadLocales 加载翻译文件
func loadLocales(bundle *config.Bundle) {
	locales := map[string]string{
		"en": "configs/i18n/locales/en.json",
		"zh": "configs/i18n/locales/zh.json",
	}
	for lang, path := range locales {
		data, err := os.ReadFile(path)
		if err != nil {
			logger.Warn(nil, "Failed to load locale", lang, "path", path, "err", err)
			continue
		}
		if err := bundle.Load(lang, data); err != nil {
			logger.Warn(nil, "Failed to parse locale", lang, "path", path, "err", err)
			continue
		}
		logger.Info(nil, "Locale loaded", "lang", lang)
	}
}
