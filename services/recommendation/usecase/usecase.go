package usecase

import (
	"github.com/NordCoder/Story/internal/entity"
	"github.com/NordCoder/Story/services/authorization/usecase"
)

// default zatychka

// todo design system that gonna fill redis with fresh categories from wiki

type RecServiceImpl struct {
	authService usecase.AuthService
}

func (i *RecServiceImpl) GetUserRec(req GetUserRecReq) (GetUserRecResp, error) {
	ww2ru := &entity.CategoryI18n{ // hardcoded must be fetched from redis
		ConceptID: 1,
		Lang:      "ru",
		Title:     "Вторая_мировая_война",
		Name:      "Вторая_мировая_война",
	}
	ww2concept := &entity.CategoryConcept{
		ID:          1,
		Key:         "world-war-ii",
		Description: "world-war-ii",
		I18ns:       []*entity.CategoryI18n{ww2ru},
	}
	return GetUserRecResp{
		recommendedCategories: []*entity.CategoryConcept{ww2concept},
	}, nil
}
