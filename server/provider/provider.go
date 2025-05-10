package provider

import (
	"horizon.com/server/horizon"
	"horizon.com/server/server/model"
	"horizon.com/server/server/repository"
)

type Provider struct {
	repo           *repository.Repository
	model          *model.Model
	authentication *horizon.HorizonAuthentication
}

func NewProvider(
	repo *repository.Repository,
	model *model.Model,
	authentication *horizon.HorizonAuthentication,

) (*Provider, error) {
	return &Provider{
		repo:           repo,
		model:          model,
		authentication: authentication,
	}, nil
}
