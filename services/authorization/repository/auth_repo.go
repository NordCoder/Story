package repository

// TODO добавить миграции

import (
	"context"
	"errors"
	"fmt"

	"github.com/NordCoder/Story/services/authorization/entity"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgxpool"
)

type authRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewAuthRepository(pool *pgxpool.Pool) UserRepository {
	return &authRepositoryImpl{db: pool}
}

func (a *authRepositoryImpl) Create(ctx context.Context, u *entity.User) (entity.UserID, error) {
	const q = `INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id`
	var id entity.UserID
	err := a.db.QueryRow(ctx, q, u.Username, u.PasswordHash).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation: // todo: i dont know why, it is not mapping(
				if pgErr.ConstraintName == "users_username_key" {
					return id, entity.ErrUsernameTaken
				}
			}
		}
		return id, fmt.Errorf("insert user: %w", err)
	}
	return id, nil
}

func (a *authRepositoryImpl) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	const q = `SELECT id, username, password_hash, created_at FROM users WHERE username=$1`
	row := a.db.QueryRow(ctx, q, username)
	var u entity.User
	if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt); err != nil {
		//if errors.Is(err, pgxpool.ErrNoRows) {
		//	return nil, entity.ErrUserNotFound
		//}
		return nil, fmt.Errorf("select user: %w", err)
	}
	return &u, nil
}
