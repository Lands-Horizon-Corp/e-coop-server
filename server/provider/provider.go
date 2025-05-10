package provider

import (
	"horizon.com/server/horizon"
	"horizon.com/server/server/model"
	"horizon.com/server/server/repository"
)

type Provider struct {
	repository     *repository.Repository
	model          *model.Model
	authentication *horizon.HorizonAuthentication
}

func NewProvider(
	repository *repository.Repository,
	model *model.Model,
	authentication *horizon.HorizonAuthentication,

) (*Provider, error) {
	return &Provider{
		repository:     repository,
		model:          model,
		authentication: authentication,
	}, nil
}
