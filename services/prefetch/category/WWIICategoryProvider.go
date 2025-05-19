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

func (W WWIICategoryProvider) GetCategory(ctx context.Context) (*entity.CategoryConcept, error) {
	ww2ru := &entity.CategoryI18n{ // hardcoded must be fetched from redis
		ConceptID: 1,
		Lang:      "ru",
		Title:     "Вторая_мировая_война",
		Name:      "Вторая_мировая_война",
	}
	ww2concept := &entity.CategoryConcept{
		ID:          1,
		Key:         "world-war-ii",
		Description: "world-war-ii",
		I18ns:       []*entity.CategoryI18n{ww2ru},
	}
	return ww2concept, nil
}

func (W WWIICategoryProvider) GetCategories(ctx context.Context) ([]*entity.CategoryConcept, error) {
	cat, err := W.GetCategory(ctx)

	if err != nil {
		return nil, err
	}

	return []*entity.CategoryConcept{cat}, nil
}

func (W WWIICategoryProvider) SetCategories(ctx context.Context, category []*entity.CategoryConcept) error {
	return errors.New("nowhere to set, this implementation doesn't support setting")
}
