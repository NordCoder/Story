package category

import (
	"context"
	"fmt"
	"math/rand"
)

// RandomCategoryProvider выбирает случайную категорию и язык из списка.
type RandomCategoryProvider struct {
	categories []CategorySelection
}

// NewRandomCategoryProvider создаёт новый RandomCategoryProvider.
func NewRandomCategoryProvider(categories []CategorySelection) *RandomCategoryProvider {
	return &RandomCategoryProvider{categories: categories}
}

// GetCategory выбирает случайную CategorySelection.
func (r *RandomCategoryProvider) GetCategory(ctx context.Context) (CategorySelection, error) {
	if len(r.categories) == 0 {
		return CategorySelection{}, fmt.Errorf("no categories available")
	}
	idx := rand.Intn(len(r.categories))
	return r.categories[idx], nil
}

func (r *RandomCategoryProvider) GetCategories(ctx context.Context) ([]CategorySelection, error) {
	return r.categories, nil
}

func (r *RandomCategoryProvider) SetCategories(ctx context.Context, categories []CategorySelection) error {
	r.categories = categories
	return nil
}
