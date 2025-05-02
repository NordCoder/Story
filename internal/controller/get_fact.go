package controller

// TODO: я не понял как найти логгер, добавить логгирование

import (
	"context"
	storypb "github.com/NordCoder/Story/generated/api/story"
	"github.com/NordCoder/Story/internal/usecase"
	"github.com/golang/protobuf/ptypes/empty"
)

func (i *implementation) GetFact(ctx context.Context, empty *empty.Empty) (*storypb.GetFactResponse, error) {
	//i.logger.Info("GetFact called")

	fact, err := i.factUseCase.GetFact(ctx, usecase.GetFactInput{})
	if err != nil {
		//i.logger.Error("failed to get fact", zap.Error(err))
		return nil, GRPCError(err)
	}

	return &storypb.GetFactResponse{
		Fact: &storypb.Fact{
			Title:   fact.Fact.Title,
			Summary: fact.Fact.Summary,
			WikiUrl: fact.Fact.SourceURL,
			ImgUrl:  fact.Fact.ImageURL,
		},
	}, nil
}
