package src

import (
	horizon_services "github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/go-playground/validator"
)

type Provider struct {
	Service *horizon_services.HorizonService
}

func NewProvider() *Provider {
	horizonService := horizon_services.NewHorizonService(horizon_services.HorizonServiceConfig{
		EnvironmentConfig: &horizon_services.EnvironmentServiceConfig{
			Path: ".env",
		},
	})
	return &Provider{
		Service: horizonService,
	}
}

func NewValidator() *validator.Validate {
	return validator.New()
}
