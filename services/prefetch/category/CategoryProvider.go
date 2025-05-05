package category

import (
	"context"
	"github.com/NordCoder/Story/internal/entity"
)

// Provider отвечает за выбор категории и языка.
type Provider interface {
	GetCategory(ctx context.Context) (*entity.CategoryConcept, error)
	GetCategories(ctx context.Context) ([]*entity.CategoryConcept, error)
	SetCategories(ctx context.Context, category []*entity.CategoryConcept) error
}
