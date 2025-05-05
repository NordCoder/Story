package config

import (
	"fmt"
	"github.com/spf13/viper"
)

const (
	httpConfigPath       = "config/http.yaml"
	loggerConfigPath     = "config/logger.yaml"
	redisConfigPath      = "config/redis.yaml"
	prefetcherConfiqPath = "config/prefetcher.yaml"
)

// -- HTTP config ---------------------------------------------------------------------------------

type HTTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	GrpcHost string `mapstructure:"grpc_host"`
	GrpcPort string `mapstructure:"grpc_port"`
	Timeouts struct {
		Read                string `mapstructure:"read"`
		Write               string `mapstructure:"write"`
		Idle                string `mapstructure:"idle"`
		ShutdownGracePeriod string `mapstructure:"shutdown_grace_period"`
	} `mapstructure:"timeouts"`
	TLS struct {
		Enabled  bool   `mapstructure:"enabled"`
		CertFile string `mapstructure:"cert_file"`
		KeyFile  string `mapstructure:"key_file"`
	} `mapstructure:"tls"`
	CORS struct {
		Enabled          bool     `mapstructure:"enabled"`
		AllowedOrigins   []string `mapstructure:"allowed_origins"`
		AllowedMethods   []string `mapstructure:"allowed_methods"`
		AllowedHeaders   []string `mapstructure:"allowed_headers"`
		AllowCredentials bool     `mapstructure:"allow_credentials"`
		MaxAge           string   `mapstructure:"max_age"`
	} `mapstructure:"cors"`
	RequestID struct {
		HeaderName        string `mapstructure:"header_name"`
		GenerateIfMissing bool   `mapstructure:"generate_if_missing"`
	} `mapstructure:"request_id"`
	Endpoints struct {
		Liveness  string `mapstructure:"liveness"`
		Readiness string `mapstructure:"readiness"`
		Metrics   string `mapstructure:"metrics"`
		Pprof     string `mapstructure:"pprof"`
	} `mapstructure:"endpoints"`
}

func NewHTTPConfig() *HTTPConfig {
	viper.SetConfigFile(httpConfigPath)
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	var httpCfg HTTPConfig
	if err := viper.UnmarshalKey("http", &httpCfg); err != nil {
		panic(fmt.Errorf("cannot parse http config: %w", err))
	}

	return &httpCfg
}

// -- LOGGER config ---------------------------------------------------------------------------------

type LoggerConfig struct {
	Level            string   `mapstructure:"level"`
	Encoding         string   `mapstructure:"encoding"`
	OutputPaths      []string `mapstructure:"output_paths"`
	ErrorOutputPaths []string `mapstructure:"error_output_paths"`
	RotatorConfig    struct {
		MaxFileSize int  `mapstructure:"max_file_size"`
		MaxBackups  int  `mapstructure:"max_backups"`
		MaxAge      int  `mapstructure:"max_age"`
		Compress    bool `mapstructure:"compress"`
	} `mapstructure:"rotation"`
}

func NewLoggerConfig() *LoggerConfig {
	viper.SetConfigFile("config/logger.yaml")
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	var loggerCfg LoggerConfig
	if err := viper.UnmarshalKey("logger", &loggerCfg); err != nil {
		panic(fmt.Errorf("cannot parse logger config: %w", err))
	}

	return &loggerCfg
}

// -- REDIS ---------------------------------------------------------------------------------------

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`

	DialTimeout  string `mapstructure:"dial_timeout"`
	ReadTimeout  string `mapstructure:"read_timeout"`
	WriteTimeout string `mapstructure:"write_timeout"`

	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
	PoolTimeout  string `mapstructure:"pool_timeout"`

	PingTimeout string `mapstructure:"ping_timeout"`
}

func NewRedisConfig() *RedisConfig {
	viper.SetConfigFile("config/redis.yaml")
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	var redisCfg RedisConfig
	if err := viper.UnmarshalKey("redis", &redisCfg); err != nil {
		panic(fmt.Errorf("cannot parse redis config: %w", err))
	}

	return &redisCfg
}
