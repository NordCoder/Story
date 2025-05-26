package config

import (
	"fmt"
	"net"
	"time"

	"github.com/spf13/viper"
)

const configFileName = "config/auth.yaml"

// AuthConfig содержит настройки авторизации и подключения к БД.
type AuthConfig struct {
	// JWT
	JWTSecret       string        `mapstructure:"jwt_secret"`
	AccessTokenTTL  time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl"`

	// Database
	DB struct {
		URL             string
		Host            string        `mapstructure:"host"`
		Port            string        `mapstructure:"port"`
		User            string        `mapstructure:"user"`
		Password        string        `mapstructure:"password"`
		Name            string        `mapstructure:"name"`
		SSLMode         string        `mapstructure:"sslmode"`
		MaxOpenConns    int           `mapstructure:"max_open_conns"`
		MaxIdleConns    int           `mapstructure:"max_idle_conns"`
		ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	} `mapstructure:"db"`
}

// NewAuthConfig загружает AuthConfig (включая DB) из auth.yaml или окружения.
func NewAuthConfig() (*AuthConfig, error) {
	v := viper.New()
	v.SetConfigFile(configFileName)
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file '%s': %w", configFileName, err)
	}

	var cfg AuthConfig
	if err := v.UnmarshalKey("auth", &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal [auth]: %w", err)
	}
	if err := v.UnmarshalKey("db", &cfg.DB); err != nil {
		return nil, fmt.Errorf("failed to unmarshal [db]: %w", err)
	}

	hostPort := net.JoinHostPort(cfg.DB.Host, cfg.DB.Port)
	cfg.DB.URL = fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		cfg.DB.User,
		cfg.DB.Password,
		hostPort,
		cfg.DB.Name,
	)

	return &cfg, nil
}
