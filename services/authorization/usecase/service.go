package usecase

import (
	"context"

	"github.com/NordCoder/Story/services/authorization/repository"
)

// default

// todo vykupit jestko authorization and implement something

type AuthService interface {
	GetUserRecInfo(ctx context.Context, req GetUserRecInfoReq) (GetUserRecInfoResp, error)
}

type GetUserRecInfoReq struct {
	UserUUID string
}

type GetUserRecInfoResp struct {
	info string
}

func NewAuthService(repository repository.AuthRepository) AuthService {
	return &AuthServiceImpl{
		repository,
	}
}
