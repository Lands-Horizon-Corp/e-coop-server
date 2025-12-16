package cable

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/server"
	"github.com/google/wire"
)

type CacheCleaner struct {
	Provider *server.Provider
}

func NewCacheCleaner(p *server.Provider) *CacheCleaner {
	return &CacheCleaner{Provider: p}
}

func (c *CacheCleaner) Clean(ctx context.Context) error {
	if err := c.Provider.Service.RunCache(ctx); err != nil {
		return err
	}
	if err := c.Provider.Service.Cache.Flush(ctx); err != nil {
		return err
	}
	return nil
}

func InitializeCacheCleaner() (*CacheCleaner, error) {
	wire.Build(
		server.NewProvider,
		NewCacheCleaner,
	)
	return nil, nil
}
