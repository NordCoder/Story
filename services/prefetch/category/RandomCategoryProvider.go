package category

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/NordCoder/Story/internal/entity"
)

// RandomCategoryProvider выбирает случайную категорию и язык из списка.
type RandomCategoryProvider struct {
	categories []entity.Category
}

// NewRandomCategoryProvider создаёт новый RandomCategoryProvider.
func NewRandomCategoryProvider(categories []entity.Category) Provider {
	return &RandomCategoryProvider{categories: categories}
}

// GetCategory выбирает случайную Selection.
func (r *RandomCategoryProvider) GetCategory(ctx context.Context) (entity.Category, error) {
	if len(r.categories) == 0 {
		return "", fmt.Errorf("no categories available")
	}
	idx := rand.Intn(len(r.categories))
	return r.categories[idx], nil
}

func (r *RandomCategoryProvider) GetCategories(ctx context.Context) ([]entity.Category, error) {
	return r.categories, nil
}

func (r *RandomCategoryProvider) SetCategories(ctx context.Context, categories []entity.Category) error {
	r.categories = categories
	return nil
}

func (r *RandomCategoryProvider) AddCategory(ctx context.Context, category entity.Category) error {
	r.categories = append(r.categories, category)
	return nil
}
