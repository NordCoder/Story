package controller

import (
	storypb "github.com/NordCoder/Story/generated/api/proto/v1"
	"github.com/NordCoder/Story/internal/usecase"
)

var _ storypb.StoryServer = (*implementation)(nil)

type implementation struct {
	factUseCase usecase.FactUseCase
}

func New(factUseCase usecase.FactUseCase) storypb.StoryServer {
	return &implementation{factUseCase: factUseCase}
}
