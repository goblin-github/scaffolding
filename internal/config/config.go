package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig    `mapstructure:"app"`
	Logger   LoggerConfig `mapstructure:"logger"`
	Database DBConfig     `mapstructure:"database"`
	JWT      JWTConfig    `mapstructure:"jwt"`
	Cache    CacheConfig  `mapstructure:"cache"`
}

type AppConfig struct {
	Port            string `mapstructure:"port"`
	Mode            string `mapstructure:"mode"`
	ReadTimeout     int    `mapstructure:"read_timeout"`
	WriteTimeout    int    `mapstructure:"write_timeout"`
	IdleTimeout     int    `mapstructure:"idle_timeout"`
	ShutdownTimeout int    `mapstructure:"shutdown_timeout"`
}

type LoggerConfig struct {
	Filename   string `mapstructure:"filename"`
	Level      string `mapstructure:"level"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
	Console    bool   `mapstructure:"console"`
}

type DBConfig struct {
	DSN             string `mapstructure:"dsn"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

type CacheConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type JWTConfig struct {
	Secret       string `mapstructure:"secret"`
	ExpiresAfter int    `mapstructure:"expires_after"` // seconds
}

func Load(filePath string) (*Config, error) {
	viper.SetConfigFile(filePath)

	viper.SetEnvPrefix("APP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config failed: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config failed: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func validate(cfg *Config) error {
	if cfg.App.Port == "" {
		return fmt.Errorf("app.port is required")
	}
	if cfg.Database.DSN == "" {
		return fmt.Errorf("database.dsn is required")
	}
	if cfg.Cache.Addr == "" {
		return fmt.Errorf("redis.addr is required")
	}
	if cfg.App.ReadTimeout <= 0 {
		cfg.App.ReadTimeout = 30
	}
	if cfg.App.WriteTimeout <= 0 {
		cfg.App.WriteTimeout = 30
	}
	if cfg.App.IdleTimeout <= 0 {
		cfg.App.IdleTimeout = 60
	}
	if cfg.App.ShutdownTimeout <= 0 {
		cfg.App.ShutdownTimeout = 10
	}
	return nil
}
