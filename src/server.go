package src

import (
	"github.com/go-playground/validator"
	horizon_services "github.com/lands-horizon/horizon-server/services"
)

type Provider struct {
	Service *horizon_services.HorizonService
}

func NewProvider() *Provider {
	horizonService := horizon_services.NewHorizonService(horizon_services.HorizonServiceConfig{
		EnvironmentConfig: &horizon_services.EnvironmentServiceConfig{
			Path: "./.env",
		},
	})
	return &Provider{
		Service: horizonService,
	}
}

func NewValidator() *validator.Validate {
	return validator.New()
}
