package repository

import (
	"context"
	"errors"

	entity2 "github.com/NordCoder/Story/internal/entity"

	"github.com/NordCoder/Story/services/authorization/entity"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotEnoughData = errors.New("not enough liked categories from user")

type RecRepository interface {
	Adjust(ctx context.Context, userID entity.UserID, category entity2.Category, delta int) error
	BulkAdjust(ctx context.Context, userID entity.UserID, categories []entity2.Category, delta int) error
	TopCategories(ctx context.Context, userID entity.UserID) ([]entity2.Category, error)
}

type recRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewRecRepository(pool *pgxpool.Pool) RecRepository {
	return &recRepositoryImpl{db: pool}
}

// Adjust increments or decrements the like count by delta (delta may be negative).
func (r recRepositoryImpl) Adjust(ctx context.Context, userID entity.UserID, category entity2.Category, delta int) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO user_category_likes(user_id, category, cnt) VALUES($1, $2, $3)
		 ON CONFLICT (user_id, category) DO UPDATE SET cnt = user_category_likes.cnt + $3`,
		userID, string(category), delta)
	return err
}

func (r recRepositoryImpl) BulkAdjust(ctx context.Context, userID entity.UserID, categories []entity2.Category, delta int) error {
	cats := make([]string, len(categories))
	for i, c := range categories {
		cats[i] = string(c)
	}
	_, err := r.db.Exec(ctx,
		`INSERT INTO user_category_likes (user_id, category, cnt)
			SELECT $1, c, $3
			FROM UNNEST($2::text[]) AS t(c)
			ON CONFLICT (user_id, category)
			DO UPDATE SET cnt = user_category_likes.cnt + $3`,
		userID, cats, delta)
	return err
}

func (r recRepositoryImpl) TopCategories(ctx context.Context, userID entity.UserID) ([]entity2.Category, error) {
	rows, err := r.db.Query(ctx,
		`SELECT category FROM user_category_likes
		 WHERE user_id = $1
		 ORDER BY cnt DESC
		 LIMIT 10`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var categories []entity2.Category
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		categories = append(categories, entity2.Category(c))
	}

	if len(categories) < 10 {
		return nil, ErrNotEnoughData
	}

	return categories, rows.Err()
}
