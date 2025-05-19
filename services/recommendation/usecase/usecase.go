package usecase

import (
	"context"

	"github.com/NordCoder/Story/internal/entity"
	"github.com/NordCoder/Story/services/authorization/controller"
	"github.com/NordCoder/Story/services/prefetch/category"
)

// default zatychka

// todo design system that gonna fill redis with fresh categories from wiki

type RecServiceImpl struct {
	authService      controller.AuthService
	categoryProvider category.Provider
}

func NewRecService(authService controller.AuthService, categoryProvider category.Provider) *RecServiceImpl {
	return &RecServiceImpl{authService, categoryProvider}
}

func (i *RecServiceImpl) GetUserRec(ctx context.Context, req GetUserRecReq) (GetUserRecResp, error) {
	cat, err := i.categoryProvider.GetCategory(ctx)

	if err != nil {
		return GetUserRecResp{}, err
	}

	cats := []*entity.CategoryConcept{cat}

	return GetUserRecResp{cats}, err
}
