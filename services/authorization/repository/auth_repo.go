package repository

import "context"

type AuthRepository interface {
	GetSomeUserInfoSomeHow(ctx context.Context, userUUID string) (string, error)
}

type authRepositoryImpl struct {
	// postgres / redis clients
}

func NewAuthRepository() AuthRepository {
	return &authRepositoryImpl{}
}

func (i *authRepositoryImpl) GetSomeUserInfoSomeHow(ctx context.Context, userUUID string) (string, error) {
	return "default", nil
}
