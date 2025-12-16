package cable

import (
	"context"

	"github.com/Lands-Horizon-Corp/e-coop-server/server"
	"github.com/fatih/color"
	"github.com/google/wire"
)

type BlockedIPAbandon struct {
	Provider *server.Provider
}

func NewBlockedIPAbandon(p *server.Provider) *BlockedIPAbandon {
	return &BlockedIPAbandon{Provider: p}
}

func (c *BlockedIPAbandon) Clear(ctx context.Context) error {
	if err := c.Provider.Service.RunCache(ctx); err != nil {
		return err
	}
	keys, err := c.Provider.Service.Cache.Keys(ctx, "blocked_ip:*")
	if err != nil {
		color.Red("Failed to get blocked IP keys: %v", err)
		return err
	}
	count := 0
	for _, key := range keys {
		if err := c.Provider.Service.Cache.Delete(ctx, key); err != nil {
			color.Red("Failed to delete key %s: %v", key, err)
		} else {
			count++
		}
	}
	color.Green("Cleared %d blocked IP entries from cache", count)
	return nil
}

func InitializeBlockedIPAbandon() (*BlockedIPAbandon, error) {
	wire.Build(
		server.NewProvider,
		NewBlockedIPAbandon,
	)
	return nil, nil
}
