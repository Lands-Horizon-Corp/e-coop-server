package handler

import (
	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
	"horizon.com/server/server/model"
	"horizon.com/server/server/provider"
	"horizon.com/server/server/publisher"
	"horizon.com/server/server/repository"
)

type Handler struct {
	authentication *horizon.HorizonAuthentication
	storage        *horizon.HorizonStorage
	provider       *provider.Provider
	repository     *repository.Repository
	model          *model.Model
	publisher      *publisher.Publisher
}

func NewHandler(
	authentication *horizon.HorizonAuthentication,
	storage *horizon.HorizonStorage,
	provider *provider.Provider,
	repository *repository.Repository,
	model *model.Model,
	publisher *publisher.Publisher,
) (*Handler, error) {
	return &Handler{
		authentication: authentication,
		storage:        storage,
		provider:       provider,
		repository:     repository,
		model:          model,
		publisher:      publisher,
	}, nil
}

func (uc *Handler) Routes(e *echo.Echo) {

}
