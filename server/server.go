package server

import (
	"github.com/Lands-Horizon-Corp/e-coop-server/services"
)

type Provider struct {
	Service *services.HorizonService
}

func NewProvider() *Provider {
	horizonService := services.NewHorizonService(services.HorizonServiceConfig{
		EnvironmentConfig: &services.EnvironmentServiceConfig{
			Path: ".env",
		},
	})
	return &Provider{
		Service: horizonService,
	}
}
