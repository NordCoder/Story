package category

// TODO: think if categories must be in-memory (I think they are not, because if there is two instances of this service they might not be synchronized)

import (
	"context"
	"github.com/NordCoder/Story/internal/entity"
)

// Provider отвечает за выбор категории и языка.
type Provider interface {
	GetCategory(ctx context.Context) (entity.CategoryConcept, error)
	GetCategories(ctx context.Context) ([]entity.CategoryConcept, error)
	SetCategories(ctx context.Context, category []entity.CategoryConcept) error
}
