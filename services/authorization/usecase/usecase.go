package usecase

import (
	"github.com/NordCoder/Story/services/authorization/repository"
)

// default

type AuthServiceImpl struct {
	authRepo repository.AuthRepository
}

func (i *AuthServiceImpl) GetUserRecInfo(req GetUserRecInfoReq) (GetUserRecInfoResp, error) {
	return GetUserRecInfoResp{}, nil
}
