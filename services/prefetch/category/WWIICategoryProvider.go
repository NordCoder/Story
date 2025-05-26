package category

import (
	"context"
	"errors"

	"github.com/NordCoder/Story/internal/entity"
)

type WWIICategoryProvider struct {
}

func NewWWIICategoryProvider() Provider {
	return &WWIICategoryProvider{}
}

func (W WWIICategoryProvider) GetCategory(ctx context.Context) (entity.Category, error) {
	ww2ru := entity.Category("Вторая_мировая_война")

	return ww2ru, nil
}

func (W WWIICategoryProvider) GetCategories(ctx context.Context) ([]entity.Category, error) {
	cat, err := W.GetCategory(ctx)

	if err != nil {
		return nil, err
	}

	return []entity.Category{cat}, nil
}

func (W WWIICategoryProvider) SetCategories(ctx context.Context, category []entity.Category) error {
	return errors.New("nowhere to set, this implementation doesn't support setting")
}

func (W WWIICategoryProvider) AddCategory(ctx context.Context, category entity.Category) error {
	panic("implement me")
}
