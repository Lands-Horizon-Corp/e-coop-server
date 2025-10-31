// Package model_logs provides database models and operations for logging functionality
package model_logs

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/src"
)

type (
	// ModelLogs represents the main model structure for logging operations
	ModelLogs struct {
		provider  *src.Provider
		Migration []any
	}
)

// NewModelLogs creates a new ModelLogs instance with the provided service provider
func NewModelLogs(provider *src.Provider) (*ModelLogs, error) {
	return &ModelLogs{
		provider: provider,
	}, nil
}

// Start initializes the logging model operations and migrations
func (c *ModelLogs) Start(_ context.Context) error {
	return nil
}
