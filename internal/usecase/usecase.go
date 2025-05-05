package usecase

import (
	"context"
	"errors"
	"github.com/NordCoder/Story/internal/entity"
	"github.com/NordCoder/Story/internal/infrastructure"
	"github.com/NordCoder/Story/internal/infrastructure/redis"
	"github.com/NordCoder/Story/services/recommendation/usecase"
)

type FactUseCase interface {
	GetFact(ctx context.Context, input GetFactInput) (GetFactOutput, error)
}

type FactUseCaseImpl struct {
	factRepo infrastructure.FactRepository

	recService usecase.RecService
}

func NewFactUseCase(factRepo infrastructure.FactRepository, recService usecase.RecService) *FactUseCaseImpl {
	return &FactUseCaseImpl{
		factRepo: factRepo,

		recService: recService,
	}
}

type GetFactInput struct{}
type GetFactOutput struct {
	Fact entity.Fact
}

func (uc *FactUseCaseImpl) GetFact(ctx context.Context, input GetFactInput) (GetFactOutput, error) {
	// category, err := uc.recService.GetUserRec(usecase.GetUserRecReq{})

	// todo get fact by category
	fact, err := uc.factRepo.PopRandom(ctx)

	if errors.Is(err, redis.ErrNotFound) || errors.Is(err, redis.ErrQueueEmpty) {
		// it is sign that prefetcher must work faster
		return GetFactOutput{
			entity.Fact{
				ID:        "-1",
				Title:     "FUN FACT",
				Summary:   "we currently don't have any facts ready...",
				ImageURL:  "",
				SourceURL: "",
				Lang:      "en",
			},
		}, nil
	}

	if err != nil {
		return GetFactOutput{}, err
	}

	return GetFactOutput{Fact: *fact}, nil
}
