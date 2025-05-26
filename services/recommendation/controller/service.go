package controller

import (
	"context"

	recpb "github.com/NordCoder/Story/generated/api/proto/v1"
	"github.com/NordCoder/Story/internal/entity"
	"github.com/NordCoder/Story/internal/logger"
	auth "github.com/NordCoder/Story/services/authorization/transport/http"
	"github.com/NordCoder/Story/services/recommendation/usecase"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

type RecService interface {
	LikeCategory(context.Context, *recpb.CategoryActionRequest) (*emptypb.Empty, error)
	UnlikeCategory(context.Context, *recpb.CategoryActionRequest) (*emptypb.Empty, error)
	GetUserRec(ctx context.Context) ([]entity.Category, error)
}

type RecServiceImpl struct {
	usecase usecase.RecUseCase
}

func NewRecService(usecase usecase.RecUseCase) RecService {
	return &RecServiceImpl{
		usecase: usecase,
	}
}

func (s *RecServiceImpl) GetUserRec(ctx context.Context) ([]entity.Category, error) {
	id, err := auth.UserIDFromCtx(ctx)
	if err != nil {
		logger.LoggerFromContext(ctx).Info("Failed to get id from context", zap.Error(err))
		return nil, err
	}
	return s.usecase.GetUserRec(ctx, id)
}

func (s *RecServiceImpl) LikeCategory(ctx context.Context, req *recpb.CategoryActionRequest) (*emptypb.Empty, error) {
	if err := req.ValidateAll(); err != nil {
		logger.LoggerFromContext(ctx).Info("LikeCategory validate fail", zap.Error(err))
		return &emptypb.Empty{}, err
	}
	id, err := auth.UserIDFromCtx(ctx)
	if err != nil {
		logger.LoggerFromContext(ctx).Info("Failed to get id from context", zap.Error(err))
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, s.usecase.LikeCategory(ctx, id, entity.Category(req.GetCategory()))
}

func (s *RecServiceImpl) UnlikeCategory(ctx context.Context, req *recpb.CategoryActionRequest) (*emptypb.Empty, error) {
	if err := req.ValidateAll(); err != nil {
		logger.LoggerFromContext(ctx).Info("UnlikeCategory validate fail", zap.Error(err))
		return &emptypb.Empty{}, err
	}
	id, err := auth.UserIDFromCtx(ctx)
	if err != nil {
		logger.LoggerFromContext(ctx).Info("Failed to get id from context", zap.Error(err))
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, s.usecase.UnlikeCategory(ctx, id, entity.Category(req.GetCategory()))
}
