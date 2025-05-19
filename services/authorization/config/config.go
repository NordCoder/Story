package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

const authConfigFileName = "auth.yaml"

// AuthConfig holds configuration for the authorization service.
type AuthConfig struct {
	// Secret key used to sign JWT access tokens.
	JWTSecret string `mapstructure:"jwt_secret"`
	// Time-to-live for access tokens (e.g., "15m").
	AccessTokenTTL time.Duration `mapstructure:"access_token_ttl"`
	// Time-to-live for refresh tokens (e.g., "168h").
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl"`
}

// NewAuthConfig loads AuthConfig from a YAML file (auth.yaml) or environment variables.
func NewAuthConfig() (*AuthConfig, error) {
	v := viper.New()
	v.SetConfigFile(authConfigFileName)
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read auth config file '%s': %w", authConfigFileName, err)
	}

	var cfg AuthConfig
	if err := v.UnmarshalKey("auth", &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal auth config: %w", err)
	}
	return &cfg, nil
}
