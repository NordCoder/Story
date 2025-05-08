package repository

import (
	"errors"
	"fmt"
	"github.com/NordCoder/Story/services/authorization/entity"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"
)

// TODO добавить миграции

import (
	"context"
	"errors"
	"fmt"
	"github.com/NordCoder/Story/services/authorization/entity"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"
)

type authRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewAuthRepository(pool *pgxpool.Pool) UserRepository {
	return &authRepositoryImpl{db: pool}
}

func (a *authRepositoryImpl) Create(ctx context.Context, u *entity.User) error {
	const q = `INSERT INTO users (id, username, password_hash) VALUES ($1, $2, $3)`
	_, err := a.db.Exec(ctx, q, u.ID, u.Username, u.PasswordHash, u.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return entity.ErrUsernameTaken
		}
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

func (a *authRepositoryImpl) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	const q = `SELECT id, username, password_hash, created_at FROM users WHERE username=$1`
	row := a.db.QueryRow(ctx, q, username)
	var u entity.User
	if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt); err != nil {
		if errors.Is(err, pgxpool.ErrNoRows) {
			return nil, entity.ErrUserNotFound
		}
		return nil, fmt.Errorf("select user: %w", err)
	}
	return &u, nil
}
