package wikipedia

import (
	"context"
	"errors"

	"github.com/NordCoder/Story/internal/entity"
)

type WikiClient interface {
	GetCategorySummaries(ctx context.Context, category entity.Category, limit int) ([]*ArticleSummary, error)
	GetSubcategories(ctx context.Context, category entity.Category, limit int) ([]entity.Category, error)
	Ping(ctx context.Context) error
}

// ErrNoPages indicates that the API returned no pages
var ErrNoPages = errors.New("wikiapi: no pages returned")

// ArticleSummary represents a Wikipedia page summary
type ArticleSummary struct {
	Title    string
	Category entity.Category
	Extract  string
	ImageURL string
	PageURL  string
}
