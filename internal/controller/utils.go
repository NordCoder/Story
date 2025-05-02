package controller

import (
	"errors"

	"github.com/NordCoder/Story/internal/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCError маппит бизнес-ошибки на gRPC-статусы.
func GRPCError(err error) error {
	switch {
	case errors.Is(err, entity.ErrNotFound):
		return status.Errorf(codes.NotFound, "resource not found")
	// можно добавлять новые случаи здесь в будущем
	default:
		return status.Errorf(codes.Internal, "internal server error")
	}
}
