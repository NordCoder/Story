package config

import (
	"github.com/spf13/viper"
	"time"
)

type PrefetcherConfig struct {
	Enabled         bool          `mapstructure:"enabled"`
	Interval        time.Duration `mapstructure:"interval"`
	BatchSize       int           `mapstructure:"batch_size"`
	MinFacts        int           `mapstructure:"min_facts"`
	PrefetchOnStart bool          `mapstructure:"prefetch_on_start"`
}

func NewPrefetcherConfig() (*PrefetcherConfig, error) {
	v := viper.New()
	v.SetConfigFile("config/prefetcher.yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg PrefetcherConfig
	if err := v.UnmarshalKey("prefetcher", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
