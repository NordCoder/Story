package category

// TODO: мейби тупой нейминг, надо изменить название пакета

import "context"

// CategorySelection описывает выбранную категорию и язык для запроса фактов.
type CategorySelection struct {
	Category string // Название категории в Википедии
	Lang     string // Язык Википедии, например "en", "ru"
}

// CategoryProvider отвечает за выбор категории и языка.
type CategoryProvider interface {
	GetCategory(ctx context.Context) (CategorySelection, error)
	GetCategories(ctx context.Context) ([]CategorySelection, error)
	SetCategories(ctx context.Context, category []CategorySelection) error
}
