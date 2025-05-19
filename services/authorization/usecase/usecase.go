package usecase

import (
	"context"

	"github.com/NordCoder/Story/services/authorization/repository"
)

// default

type AuthServiceImpl struct {
	authRepo repository.AuthRepository
}

func (i *AuthServiceImpl) GetUserRecInfo(ctx context.Context, req GetUserRecInfoReq) (GetUserRecInfoResp, error) {
	return GetUserRecInfoResp{}, nil
}
