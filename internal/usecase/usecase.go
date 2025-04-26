package usecase

import (
	"context"
	"github.com/NordCoder/Story/internal/entity"
	"github.com/NordCoder/Story/internal/infrastructure"
)

type FactUseCase interface {
	GetFact(ctx context.Context, input GetFactInput) (GetFactOutput, error)
}

type FactUseCaseImpl struct {
	factRepo           infrastructure.FactRepository
	factRepoTransactor infrastructure.Transactor
}

func NewFactUseCase(factRepo infrastructure.FactRepository) *FactUseCaseImpl {
	return &FactUseCaseImpl{factRepo: factRepo}
}

type GetFactInput struct{}
type GetFactOutput struct {
	Fact entity.Fact
}

func (uc *FactUseCaseImpl) GetFact(ctx context.Context, input GetFactInput) (GetFactOutput, error) {
	var fact *entity.Fact

	err := uc.factRepoTransactor.WithTx(ctx, func(ctx context.Context) error {
		var err error

		fact, err = uc.factRepo.PopRandom(ctx)

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return GetFactOutput{}, err
	}

	return GetFactOutput{Fact: *fact}, nil
}
