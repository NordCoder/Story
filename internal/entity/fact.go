package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// FactID — уникальный идентификатор факта (UUID v4).
type FactID string

// NewFactID генерирует новый уникальный идентификатор факта (UUID v4).
func NewFactID() FactID {
	return FactID(uuid.New().String())
}

// Fact — единица контента, которую мы доставляем пользователю.
type Fact struct {
	ID        FactID    // uuid, создаётся при fetch
	Category  Category  // категория данного факта
	Title     string    // заголовок статьи
	Summary   string    // обрезанный текст (≤280 символов)
	ImageURL  string    // URL изображения-миниатюры (nullable)
	SourceURL string    // полный URL на статью в Википедии
	Lang      string    // языковой код статьи, например "ru"
	FetchedAt time.Time // время получения от Wikipedia API
}

// ErrFactNotFound возвращается, когда факт с данным ID не найден.
var ErrFactNotFound = errors.New("fact not found")
