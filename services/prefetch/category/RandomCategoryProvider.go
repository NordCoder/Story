package category

import (
	"context"
	"fmt"
	"github.com/NordCoder/Story/internal/entity"
	"math/rand"
)

// RandomCategoryProvider выбирает случайную категорию и язык из списка.
type RandomCategoryProvider struct {
	categories []*entity.CategoryConcept
}

// NewRandomCategoryProvider создаёт новый RandomCategoryProvider.
func NewRandomCategoryProvider(categories []*entity.CategoryConcept) Provider {
	return &RandomCategoryProvider{categories: categories}
}

// GetCategory выбирает случайную Selection.
func (r *RandomCategoryProvider) GetCategory(ctx context.Context) (*entity.CategoryConcept, error) {
	if len(r.categories) == 0 {
		return &entity.CategoryConcept{}, fmt.Errorf("no categories available")
	}
	idx := rand.Intn(len(r.categories))
	return r.categories[idx], nil
}

func (r *RandomCategoryProvider) GetCategories(ctx context.Context) ([]*entity.CategoryConcept, error) {
	return r.categories, nil
}

func (r *RandomCategoryProvider) SetCategories(ctx context.Context, categories []*entity.CategoryConcept) error {
	r.categories = categories
	return nil
}
