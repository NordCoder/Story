package infrastructure

import (
	"context"

	"github.com/NordCoder/Story/internal/entity"
)

// FactRepository описывает хранилище фактов.
//
//	– Save сохраняет новый факт; перезапись по тому же ID не происходит.
//	– GetByID возвращает факт по ID или ErrNotFound.
//	– PopRandom извлекает и удаляет один случайный ID из очереди, возвращая весь факт.
type FactRepository interface {
	Save(ctx context.Context, f *entity.Fact) error
	GetByID(ctx context.Context, id entity.FactID) (*entity.Fact, error)
	PopRandom(ctx context.Context) (*entity.Fact, error)
}

type FetchClient interface {
	GetSummary(ctx context.Context, dto *FetchRequestDTO) (*FetchResponseDTO, error)
}

type PreFetcher interface {
	FetchToRedis(ctx context.Context, dto *FetchRequestDTO) (*FetchResponseDTO, error)
}

type FetchResponseDTO struct {
	Title     string `json:"title"`
	Extract   string `json:"extract"`
	Thumbnail struct {
		Source string `json:"source"`
	} `json:"thumbnail"`
	ContentURLs struct {
		Desktop struct {
			Page string `json:"page"`
		} `json:"desktop"`
	} `json:"content_urls"`
}

type FetchRequestDTO struct {
}
