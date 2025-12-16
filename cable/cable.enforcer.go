package cable

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/fatih/color"
)

type BlocklistEnforcer struct {
	Provider *server.Provider
	Core     *core.Core
}

func NewBlocklistEnforcer(p *server.Provider, c *core.Core) *BlocklistEnforcer {
	return &BlocklistEnforcer{Provider: p, Core: c}
}

func (b *BlocklistEnforcer) Enforce(ctx context.Context) error {
	if err := b.Provider.Service.RunCache(ctx); err != nil {
		return err
	}
	return b.Provider.Service.Security.Firewall(ctx, func(ip, host string) {
		cacheKey := "blocked_ip:" + ip
		timestamp := float64(time.Now().Unix())

		if err := b.Provider.Service.Cache.ZAdd(ctx, "blocked_ips_registry", timestamp, ip); err != nil {
			color.Red("Failed to add IP %s to registry: %v", ip, err)
		}

		if err := b.Provider.Service.Cache.Set(ctx, cacheKey, host, 60*24*time.Hour); err != nil {
			color.Red("Failed to cache IP %s: %v", ip, err)
		}
		color.Yellow("Cached blocked IP %s from host %s", ip, host)
	})
}

// func InitializeBlocklistEnforcer() (*BlocklistEnforcer, error) {
// 	wire.Build(
// 		server.NewProvider,
// 		core.NewCore,
// 		NewBlocklistEnforcer,
// 	)
// 	return nil, nil
// }
