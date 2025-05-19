package usecase

import (
	"context"

	"github.com/NordCoder/Story/internal/entity"
)

type RecService interface {
	GetUserRec(ctx context.Context, req GetUserRecReq) (GetUserRecResp, error)
}

type GetUserRecReq struct {
	userUUID string
}

type GetUserRecResp struct {
	recommendedCategories []*entity.CategoryConcept
}
