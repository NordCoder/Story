package usecase

import (
	"github.com/NordCoder/Story/internal/entity"
)

type RecService interface {
	GetUserRec(req GetUserRecReq) (GetUserRecResp, error)
}

type GetUserRecReq struct {
	userUUID string
}

type GetUserRecResp struct {
	recommendedCategories []*entity.CategoryConcept
}
