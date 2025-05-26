package category

import (
	"context"

	"github.com/NordCoder/Story/internal/entity"
)

// Provider отвечает за выбор категории и языка.
type Provider interface {
	GetCategory(ctx context.Context) (entity.Category, error)
	GetCategories(ctx context.Context) ([]entity.Category, error)
	SetCategories(ctx context.Context, category []entity.Category) error
	AddCategory(ctx context.Context, category entity.Category) error
}
