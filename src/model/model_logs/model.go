package model_logs

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/src"
)

type (
	ModelLogs struct {
		provider  *src.Provider
		Migration []any
	}
)

func NewModelLogs(provider *src.Provider) (*ModelLogs, error) {
	return &ModelLogs{
		provider: provider,
	}, nil
}
func (c *ModelLogs) Start(context context.Context) error {
	return nil
}
