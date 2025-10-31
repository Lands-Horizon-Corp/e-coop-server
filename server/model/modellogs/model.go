// Package modellogs provides database models and operations for logging functionality
package modellogs

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/server"
)

type (
	// ModelLogs represents the main model structure for logging operations
	ModelLogs struct {
		provider  *server.Provider
		Migration []any
	}
)

// NewModellogs creates a new ModelLogs instance with the provided service provider
func NewModellogs(provider *server.Provider) (*ModelLogs, error) {
	return &ModelLogs{
		provider: provider,
	}, nil
}

// Start initializes the logging model operations and migrations
func (c *ModelLogs) Start(_ context.Context) error {
	return nil
}
