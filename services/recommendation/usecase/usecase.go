package usecase

import (
	"context"

	entity2 "github.com/NordCoder/Story/internal/entity"

	"github.com/NordCoder/Story/internal/logger"
	"github.com/NordCoder/Story/services/authorization/entity"
	"github.com/NordCoder/Story/services/recommendation/repository"
	"go.uber.org/zap"
)

type propagateTask struct {
	userID   entity.UserID
	category entity2.Category
	depth    int
}

// buffered channel for background propagation tasks
var propagateCh = make(chan propagateTask, 1000)

// default zatychka

// todo design system that gonna fill redis with fresh categories from wiki

type RecUseCase interface {
	LikeCategory(context.Context, entity.UserID, entity2.Category) error
	UnlikeCategory(context.Context, entity.UserID, entity2.Category) error
	GetUserRec(ctx context.Context, id entity.UserID) ([]entity2.Category, error)
}

type RecUseCaseImpl struct {
	recRepo repository.RecRepository
}

func NewRecUseCase(recRepo repository.RecRepository) RecUseCase {
	return &RecUseCaseImpl{
		recRepo: recRepo,
	}
}

func (r RecUseCaseImpl) LikeCategory(ctx context.Context, id entity.UserID, category entity2.Category) error {
	logger.LoggerFromContext(ctx).Info("LikeCategory usecase starts", zap.String("category: ", string(category)))
	err := r.recRepo.Adjust(ctx, id, category, 1)
	if err != nil {
		logger.LoggerFromContext(ctx).Error("Incr Error", zap.Error(err))
		return err
	}
	select {
	case propagateCh <- propagateTask{userID: id, category: category, depth: 1}:
	default:
		logger.LoggerFromContext(ctx).Warn("propagateCh full, dropping propagation task", zap.String("category", string(category)), zap.String("user_id", string(id)))
	}
	return nil
}

func (r RecUseCaseImpl) UnlikeCategory(ctx context.Context, id entity.UserID, category entity2.Category) error {
	logger.LoggerFromContext(ctx).Info("UnlikeCategory usecase starts", zap.String("category: ", string(category)))
	err := r.recRepo.Adjust(ctx, id, category, -1)
	if err != nil {
		logger.LoggerFromContext(ctx).Error("Decrement Error", zap.Error(err))
		return err
	}
	return nil
}

func (r RecUseCaseImpl) GetUserRec(ctx context.Context, id entity.UserID) ([]entity2.Category, error) {
	logger.LoggerFromContext(ctx).Info("GetUserRec usecase starts")
	categories, err := r.recRepo.TopCategories(ctx, id)
	if err != nil {
		logger.LoggerFromContext(ctx).Error("TopCategories Error", zap.Error(err))
		return nil, err
	}
	return categories, nil
}
