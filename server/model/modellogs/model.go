package modellogs

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/server"
)

type (
	ModelLogs struct {
		provider  *server.Provider
		Migration []any
	}
)

func NewModellogs(provider *server.Provider) (*ModelLogs, error) {
	return &ModelLogs{
		provider: provider,
	}, nil
}

func (c *ModelLogs) Start(_ context.Context) error {
	return nil
}
