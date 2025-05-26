package controller

import (
	"context"
	storypb "github.com/NordCoder/Story/generated/api/proto/v1"
	"github.com/NordCoder/Story/internal/logger"
	"github.com/NordCoder/Story/internal/usecase"
	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"
)

func (i *implementation) GetFact(ctx context.Context, empty *empty.Empty) (*storypb.GetFactResponse, error) {
	logger.LoggerFromContext(ctx).Info("getting fact")

	fact, err := i.factUseCase.GetFact(ctx, usecase.GetFactInput{})
	if err != nil {
		logger.LoggerFromContext(ctx).Error("failed to get fact", zap.Error(err))
		return nil, GRPCError(err)
	}

	return &storypb.GetFactResponse{
		Fact: &storypb.Fact{
			Title:    fact.Fact.Title,
			Category: string(fact.Fact.Category),
			Summary:  fact.Fact.Summary,
			WikiUrl:  fact.Fact.SourceURL,
			ImgUrl:   fact.Fact.ImageURL,
		},
	}, nil
}
