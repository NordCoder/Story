package config

import (
	"time"

	"github.com/spf13/viper"
)

const prefetcherConfigPath = "prefetch.yaml"

type PrefetcherConfig struct {
	Enabled         bool          `mapstructure:"enabled"`
	Interval        time.Duration `mapstructure:"interval"`
	BatchSize       int           `mapstructure:"batch_size"`
	MinFacts        int           `mapstructure:"min_facts"`
	PrefetchOnStart bool          `mapstructure:"prefetch_on_start"`
}

func NewPrefetcherConfig() (*PrefetcherConfig, error) {
	viper.SetConfigFile(prefetcherConfigPath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg PrefetcherConfig
	if err := viper.UnmarshalKey("prefetcher", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
