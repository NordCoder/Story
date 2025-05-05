package wikipedia

import (
	"context"
	"errors"
)

type WikiClient interface {
	GetCategorySummaries(ctx context.Context, category string, limit int) ([]*ArticleSummary, error)
	Ping(ctx context.Context) error
}

// ErrNoPages indicates that the API returned no pages
var ErrNoPages = errors.New("wikiapi: no pages returned")

// ArticleSummary represents a Wikipedia page summary
type ArticleSummary struct {
	Title    string
	Extract  string
	ImageURL string
	PageURL  string
}
