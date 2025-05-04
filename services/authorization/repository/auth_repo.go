package repository

type AuthRepository interface {
	GetSomeUserInfoSomeHow(userUUID string) (string, error)
}

type authRepositoryImpl struct {
	// postgres / redis clients
}

func NewAuthRepository() AuthRepository {
	return &authRepositoryImpl{}
}

func (i *authRepositoryImpl) GetSomeUserInfoSomeHow(userUUID string) (string, error) {
	return "default", nil
}
