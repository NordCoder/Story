package entity

import (
	"errors"
	"time"
)

// UserID uniquely identifies a user (UUID v4).
type UserID string

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrUsernameTaken   = errors.New("username already taken")
	ErrInvalidPassword = errors.New("invalid password")
)

// User represents a registered account in the system.
type User struct {
	ID           UserID    // unique user ID
	Username     string    // login username, unique
	PasswordHash string    // hashed password (bcrypt or argon2id)
	CreatedAt    time.Time // timestamp of registration
}
