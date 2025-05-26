package usecase

import (
	"context"
	"time"

	apiv1 "github.com/NordCoder/Story/generated/api/proto/v1"
	"github.com/NordCoder/Story/internal/logger"
	"github.com/NordCoder/Story/services/authorization/config"
	"github.com/NordCoder/Story/services/authorization/entity"
	"github.com/NordCoder/Story/services/authorization/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// AuthUseCase defines business logic operations using domain types and returns proto responses.
type AuthUseCase interface {
	Register(ctx context.Context, username, password string) (*apiv1.RegisterResponse, error)
	Login(ctx context.Context, username, password string) (*apiv1.LoginResponse, error)
	Refresh(ctx context.Context, refreshToken string) (*apiv1.RefreshResponse, error)
	Logout(ctx context.Context, refreshToken string) (*apiv1.LogoutResponse, error)
}

type AuthUseCaseImpl struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	cfg         *config.AuthConfig
}

// generateAccessToken creates a JWT access token for the given user ID.
func (u *AuthUseCaseImpl) generateAccessToken(userID entity.UserID) (tokenStr string, expiresIn int64, err error) {
	now := time.Now()
	expiresAt := now.Add(u.cfg.AccessTokenTTL).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"iat": now.Unix(),
		"exp": expiresAt,
	})
	tokenStr, err = token.SignedString([]byte(u.cfg.JWTSecret))
	if err != nil {
		return "", 0, err
	}
	return tokenStr, int64(u.cfg.AccessTokenTTL.Seconds()), nil
}

func NewAuthUseCaseImpl(userRepo repository.UserRepository, sessionRepo repository.SessionRepository, cfg *config.AuthConfig) *AuthUseCaseImpl {
	return &AuthUseCaseImpl{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		cfg:         cfg,
	}
}

func (u *AuthUseCaseImpl) Register(ctx context.Context, username, password string) (*apiv1.RegisterResponse, error) {
	logger.LoggerFromContext(ctx).Info("Registering user")

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.LoggerFromContext(ctx).Info("failed to hash password", zap.Error(err))
		return nil, err
	}
	user := &entity.User{
		Username:     username,
		PasswordHash: string(hashed),
	}
	if user.ID, err = u.userRepo.Create(ctx, user); err != nil {
		logger.LoggerFromContext(ctx).Info("failed to create user", zap.Error(err))
		return nil, err
	}
	return &apiv1.RegisterResponse{UserId: string(user.ID)}, nil
}

func (u *AuthUseCaseImpl) Login(ctx context.Context, username, password string) (*apiv1.LoginResponse, error) {
	logger.LoggerFromContext(ctx).Info("Logging in user")
	user, err := u.userRepo.FindByUsername(ctx, username)
	if err != nil {
		logger.LoggerFromContext(ctx).Info("failed to find user", zap.Error(err))
		return nil, err
	}
	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		logger.LoggerFromContext(ctx).Info("failed to compare password", zap.Error(err))
		return nil, entity.ErrInvalidPassword
	}

	accessStr, expiresIn, err := u.generateAccessToken(user.ID)
	if err != nil {
		logger.LoggerFromContext(ctx).Info("failed to generate access token", zap.Error(err))
		return nil, err
	}

	refreshToken := uuid.New().String()
	if err = u.sessionRepo.SaveRefreshToken(ctx, refreshToken, user.ID, 0); err != nil {
		logger.LoggerFromContext(ctx).Info("failed to save refresh token", zap.Error(err))
		return nil, err
	}
	return &apiv1.LoginResponse{
		AccessToken:  accessStr,
		ExpiresIn:    expiresIn,
		RefreshToken: refreshToken,
	}, nil
}

func (u *AuthUseCaseImpl) Refresh(ctx context.Context, oldRefreshToken string) (*apiv1.RefreshResponse, error) {
	logger.LoggerFromContext(ctx).Info("Refreshing user")

	userID, err := u.sessionRepo.GetUserIDByRefreshToken(ctx, oldRefreshToken)
	if err != nil {
		logger.LoggerFromContext(ctx).Info("failed to find user", zap.Error(err))
		return nil, err
	}

	newRefreshToken := uuid.New().String()

	if err = u.sessionRepo.RotateRefreshToken(ctx, oldRefreshToken, newRefreshToken, userID, u.cfg.RefreshTokenTTL); err != nil {
		logger.LoggerFromContext(ctx).Info("failed to rotate refresh token")
		return nil, err
	}

	accessStr, expiresIn, err := u.generateAccessToken(userID)
	if err != nil {
		logger.LoggerFromContext(ctx).Error("failed to generate access token", zap.Error(err))
		return nil, err
	}

	return &apiv1.RefreshResponse{
		AccessToken:  accessStr,
		ExpiresIn:    expiresIn,
		RefreshToken: newRefreshToken,
	}, nil
}

// todo: make sense to delete from bd also

func (u *AuthUseCaseImpl) Logout(ctx context.Context, refreshToken string) (*apiv1.LogoutResponse, error) {
	logger.LoggerFromContext(ctx).Info("Logging out user")
	if err := u.sessionRepo.DeleteRefreshToken(ctx, refreshToken); err != nil {
		logger.LoggerFromContext(ctx).Error("failed to delete refresh token", zap.Error(err))
		return nil, err
	}
	return &apiv1.LogoutResponse{Success: true}, nil
}
