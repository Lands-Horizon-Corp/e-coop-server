package core

import (
	"github.com/Lands-Horizon-Corp/e-coop-server/server"
	"github.com/Lands-Horizon-Corp/e-coop-server/services/registry"
)

type (
	Core struct {
		provider  *server.Provider
		Migration []any

		MediaManager registry.Registry[Media, MediaResponse, MediaRequest]
	}
)

func NewCore(provider *server.Provider) (*Core, error) {
	return &Core{
		provider: provider,
	}, nil
}

func (c *Core) Start() error {
	c.media()
	return nil
}
