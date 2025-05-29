package category

import (
	"context"
	"fmt"
	"github.com/NordCoder/Story/internal/entity"
	"math/rand"
)

type DefaultProvider struct {
	categories []entity.Category
}

func NewDefaultProvider() *DefaultProvider {
	return &DefaultProvider{
		categories: []entity.Category{"Втоарая_мировая_война",
			"Операции_и_сражения_Второй_мировой_войны",
			"Участники_Второй_мировой_войны",
			"Политика_во_Второй_мировой_войне",
			"Театры_военных_действий_Второй_мировой_войны",
			"Сопротивленческие_движения_Во_второй_мировой_войне",
			"Память_о_Второй_мировой_войне",
			"Потери_во_Второй_мировой_войне",
			"Награды_Второй_мировой_войны",
			"Хронология_Второй_мировой_войны",
			"Вторая_мировая_война_на_море"},
	}
}

func (p *DefaultProvider) GetCategory(_ context.Context) (entity.Category, error) {
	if len(p.categories) == 0 {
		return "", fmt.Errorf("no categories available")
	}
	idx := rand.Intn(len(p.categories))
	return p.categories[idx], nil
}

func (p *DefaultProvider) GetCategories(_ context.Context) ([]entity.Category, error) {
	return p.categories, nil
}

func (p *DefaultProvider) SetCategories(_ context.Context, categories []entity.Category) error {
	p.categories = categories
	return nil
}

func (p *DefaultProvider) AddCategory(_ context.Context, category entity.Category) error {
	p.categories = append(p.categories, category)
	return nil
}
