package entity

import (
	"errors"
	"time"
)

// FactID — уникальный идентификатор факта (UUID v4).
type FactID string

// Fact — единица контента, которую мы доставляем пользователю.
type Fact struct {
	ID        FactID    // uuid, создаётся при fetch
	Title     string    // заголовок статьи
	Summary   string    // обрезанный текст (≤280 символов)
	ImageURL  string    // URL изображения-миниатюры (nullable)
	SourceURL string    // полный URL на статью в Википедии
	Lang      string    // языковой код статьи, например "ru"
	FetchedAt time.Time // время получения от Wikipedia API
}

// ErrNotFound возвращается, когда факт с данным ID не найден.
var ErrNotFound = errors.New("fact not found")
