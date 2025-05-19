package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/NordCoder/Story/config"
	"github.com/go-redis/redis/v8"
)

func NewRedisClient() (*redis.Client, error) {
	cfg := config.NewRedisConfig()

	dialTO, _ := time.ParseDuration(cfg.DialTimeout)
	readTO, _ := time.ParseDuration(cfg.ReadTimeout)
	writeTO, _ := time.ParseDuration(cfg.WriteTimeout)
	poolTO, _ := time.ParseDuration(cfg.PoolTimeout)

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,

		DialTimeout:  dialTO,
		ReadTimeout:  readTO,
		WriteTimeout: writeTO,

		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		PoolTimeout:  poolTO,
	})

	pingTO, _ := time.ParseDuration(cfg.PingTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), pingTO)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}
	return rdb, nil
}
