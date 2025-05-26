package controller

// TODO разбить по файлам мб
// todo: error mapping mb

import (
	"context"

	storypb "github.com/NordCoder/Story/generated/api/proto/v1"
	"github.com/NordCoder/Story/internal/logger"
	"github.com/NordCoder/Story/services/authorization/usecase"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService interface {
	Register(context.Context, *storypb.RegisterRequest) (*storypb.RegisterResponse, error)
	Login(context.Context, *storypb.LoginRequest) (*storypb.LoginResponse, error)
	Refresh(context.Context, *storypb.RefreshRequest) (*storypb.RefreshResponse, error)
	Logout(context.Context, *storypb.LogoutRequest) (*storypb.LogoutResponse, error)
}

type AuthServiceImpl struct {
	usecase usecase.AuthUseCase
}

func NewAuthService(usecase usecase.AuthUseCase) AuthService {
	return &AuthServiceImpl{
		usecase: usecase,
	}
}

func (a AuthServiceImpl) Register(ctx context.Context, request *storypb.RegisterRequest) (*storypb.RegisterResponse, error) {
	if err := request.ValidateAll(); err != nil {
		logger.LoggerFromContext(ctx).Info("Invalid request", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return a.usecase.Register(ctx, request.Username, request.Password)
}

func (a AuthServiceImpl) Login(ctx context.Context, request *storypb.LoginRequest) (*storypb.LoginResponse, error) {
	if err := request.ValidateAll(); err != nil {
		logger.LoggerFromContext(ctx).Info("Invalid request", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return a.usecase.Login(ctx, request.Username, request.Password)
}

func (a AuthServiceImpl) Refresh(ctx context.Context, request *storypb.RefreshRequest) (*storypb.RefreshResponse, error) {
	if err := request.ValidateAll(); err != nil {
		logger.LoggerFromContext(ctx).Info("Invalid request", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return a.usecase.Refresh(ctx, request.RefreshToken)
}

func (a AuthServiceImpl) Logout(ctx context.Context, request *storypb.LogoutRequest) (*storypb.LogoutResponse, error) {
	if err := request.ValidateAll(); err != nil {
		logger.LoggerFromContext(ctx).Info("Invalid request", zap.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return a.usecase.Logout(ctx, request.RefreshToken)
}
