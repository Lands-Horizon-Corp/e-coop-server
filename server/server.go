// Package server provides core server components and provider initialization for the e-cooperative application
package server

import (
	"github.com/Lands-Horizon-Corp/e-coop-server/services"
	"github.com/go-playground/validator"
)

// Provider contains the main service provider for the application
type Provider struct {
	Service *services.HorizonService
}

// NewProvider creates a new Provider instance with initialized horizon service
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

// NewValidator creates a new validator instance for request validation
func NewValidator() *validator.Validate {
	return validator.New()
}
