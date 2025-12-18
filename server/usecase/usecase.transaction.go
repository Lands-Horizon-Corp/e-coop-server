package usecase

import (
	"github.com/Lands-Horizon-Corp/e-coop-server/server"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
)

type UsecaseService struct {
	model    *core.Core
	provider *server.Provider
}

func NewUsecaseService(
	model *core.Core,
	provider *server.Provider,
) (*UsecaseService, error) {
	return &UsecaseService{
		model:    model,
		provider: provider,
	}, nil
}
