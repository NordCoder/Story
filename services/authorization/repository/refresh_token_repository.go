package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/NordCoder/Story/services/authorization/entity"
	"github.com/go-redis/redis/v8"
)

// RefreshTokenRepository manages storage of refresh tokens in Redis.
// Keys are formatted with a prefix and the token string.
type RefreshTokenRepository struct {
	client    *redis.Client
	ttl       time.Duration
	keyPrefix string
}

// key formats the Redis key for a given token.
func (r *RefreshTokenRepository) key(token string) string {
	return fmt.Sprintf(r.keyPrefix, token)
}

// save writes the token->userID mapping with TTL into Redis.
func (r *RefreshTokenRepository) save(ctx context.Context, cli redis.Cmdable, token string, userID entity.UserID, ttl time.Duration) {
	cli.Set(ctx, r.key(token), string(userID), ttl)
}

// del removes the token key in Redis.
func (r *RefreshTokenRepository) del(ctx context.Context, cli redis.Cmdable, token string) {
	cli.Del(ctx, r.key(token))
}

// NewRefreshTokenRepository creates a new RefreshTokenRepository.
// ttl is the default expiration for refresh tokens.
// By default, keyPrefix="auth:refresh:%s" but can be overridden with WithKeyPrefix.
func NewRefreshTokenRepository(client *redis.Client, ttl time.Duration, opts ...Option) *RefreshTokenRepository {
	repo := &RefreshTokenRepository{
		client:    client,
		ttl:       ttl,
		keyPrefix: "auth:refresh:%s",
	}
	for _, o := range opts {
		o(repo)
	}
	return repo
}

// Option configures RefreshTokenRepository behavior.
type Option func(*RefreshTokenRepository)

// WithKeyPrefix sets a custom key prefix (format string) for tokens.
func WithKeyPrefix(prefix string) Option {
	return func(r *RefreshTokenRepository) { r.keyPrefix = prefix }
}

// SaveRefreshToken stores a refresh token with its associated user ID and TTL.
func (r *RefreshTokenRepository) SaveRefreshToken(ctx context.Context, token string, userID entity.UserID, ttl time.Duration) error {
	res := r.client.Set(ctx, r.key(token), string(userID), ttl)
	return res.Err()
}

// GetUserIDByRefreshToken retrieves the user ID for a given refresh token.
// Returns ErrRefreshNotFound if the token does not exist or is expired.
func (r *RefreshTokenRepository) GetUserIDByRefreshToken(ctx context.Context, token string) (entity.UserID, error) {
	key := fmt.Sprintf(r.keyPrefix, token)
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrRefreshNotFound
		}
		return "", err
	}
	return entity.UserID(val), nil
}

// DeleteRefreshToken removes a refresh token from Redis.
func (r *RefreshTokenRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	res := r.client.Del(ctx, r.key(token))
	return res.Err()
}

// RotateRefreshToken atomically replaces an old refresh token with a new one for a user.
func (r *RefreshTokenRepository) RotateRefreshToken(ctx context.Context, oldToken, newToken string, userID entity.UserID, ttl time.Duration) error {
	pipe := r.client.TxPipeline()
	r.del(ctx, pipe, oldToken)
	r.save(ctx, pipe, newToken, userID, ttl)
	_, err := pipe.Exec(ctx)
	return err
}

// ErrRefreshNotFound is returned when a refresh token is not found in storage.
var ErrRefreshNotFound = redis.Nil
