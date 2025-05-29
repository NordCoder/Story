package category

import (
	"context"
	"github.com/NordCoder/Story/internal/entity"
)

type StackProvider struct {
	categories []entity.Category
}

func NewStackProvider() *StackProvider {
	return &StackProvider{
		categories: make([]entity.Category, 0),
	}
}

func (s *StackProvider) GetCategory(_ context.Context) (entity.Category, error) {
	cat := s.categories[len(s.categories)-1]
	s.categories = s.categories[:len(s.categories)-1]
	return cat, nil
}

func (s *StackProvider) GetCategories(_ context.Context) ([]entity.Category, error) {
	return s.categories, nil
}

func (s *StackProvider) SetCategories(_ context.Context, categories []entity.Category) error {
	s.categories = categories
	return nil
}

func (s *StackProvider) AddCategory(_ context.Context, category entity.Category) error {
	s.categories = append(s.categories, category)
	return nil
}
