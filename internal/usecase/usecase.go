package usecase

import (
	"context"
	"errors"
	"math/rand"

	"github.com/NordCoder/Story/internal/entity"
	"github.com/NordCoder/Story/internal/infrastructure"
	"github.com/NordCoder/Story/internal/logger"
	"github.com/NordCoder/Story/services/recommendation/controller"
	"go.uber.org/zap"
)

var (
	ErrNotFound = errors.New("no fact found")
)

type FactUseCase interface {
	GetFact(ctx context.Context, input GetFactInput) (GetFactOutput, error)
}

type FactUseCaseImpl struct {
	factRepo infrastructure.FactRepository

	recService controller.RecService
}

func NewFactUseCase(factRepo infrastructure.FactRepository, recService controller.RecService) *FactUseCaseImpl {
	return &FactUseCaseImpl{
		factRepo: factRepo,

		recService: recService,
	}
}

type GetFactInput struct{}
type GetFactOutput struct {
	Fact entity.Fact
}

func returnDefault() GetFactOutput {
	return GetFactOutput{
		entity.Fact{
			ID:        "-1",
			Title:     "FUN FACT",
			Summary:   "we currently don't have any facts ready...",
			ImageURL:  "https://upload.wikimedia.org/wikipedia/commons/3/37/Sad-face.png",
			SourceURL: "",
			Lang:      "en",
		},
	}
}

func (uc *FactUseCaseImpl) GetFact(ctx context.Context, input GetFactInput) (GetFactOutput, error) {
	cats, err := uc.recService.GetUserRec(ctx)

	r := rand.Float64()
	byCategory := err == nil && len(cats) > 0 && r < 0.6

	var fact *entity.Fact
	var category entity.Category

	if byCategory {
		category = cats[0]
		logger.LoggerFromContext(ctx).Info("GetFact: trying by category",
			zap.String("category", string(category)),
			zap.Float64("rand", r),
		)

		facts, err2 := uc.factRepo.GetByCategory(ctx, category, 10)
		if err2 == nil && len(facts) > 0 {
			fact = facts[rand.Intn(len(facts))]
		} else {
			zap.L().Warn("GetFact: empty category, falling back to random", zap.Error(err2), zap.String("category", string(category)))
			fact, err = uc.factRepo.PopRandom(ctx)

			zap.L().Warn("fact", zap.Bool("fact is nil", fact == nil))

			if err != nil || fact == nil {
				return returnDefault(), nil
			}

		}
	} else {
		zap.L().Info("GetFact: random path", zap.Float64("rand", r))
		fact, err = uc.factRepo.PopRandom(ctx)

		zap.L().Warn("fact", zap.Bool("fact is nil", fact == nil))

		if err != nil || fact == nil {
			return returnDefault(), nil
		}
	}

	return GetFactOutput{Fact: *fact}, nil
}
