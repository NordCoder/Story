package repository

import (
	"context"
	"time"

	"github.com/NordCoder/Story/services/authorization/entity"
)

// UserRepository defines operations for managing user accounts.
type UserRepository interface {
	// Create inserts a new user into the system.
	Create(ctx context.Context, u *entity.User) error

	// FindByUsername retrieves a user by username.
	// Returns ErrUserNotFound if no such user exists.
	FindByUsername(ctx context.Context, username string) (*entity.User, error)
}

// SessionRepository defines operations for managing user sessions via refresh tokens.
type SessionRepository interface {
	// SaveRefreshToken stores a refresh token for a user with a given TTL.
	SaveRefreshToken(ctx context.Context, token string, userID entity.UserID, ttl time.Duration) error

	// DeleteRefreshToken invalidates a refresh token.
	DeleteRefreshToken(ctx context.Context, token string) error

	// GetUserIDByRefreshToken looks up the user associated with a refresh token.
	// Returns ErrRefreshNotFound if the token is missing or expired.
	GetUserIDByRefreshToken(ctx context.Context, token string) (entity.UserID, error)

	// RotateRefreshToken atomically replaces an old refresh token with a new one for a user.
	RotateRefreshToken(ctx context.Context, oldToken, newToken string, userID entity.UserID, ttl time.Duration) error
}
