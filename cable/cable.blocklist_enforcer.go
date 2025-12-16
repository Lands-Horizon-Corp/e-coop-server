package cable

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/fatih/color"
	"github.com/google/wire"
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

func InitializeBlocklistEnforcer() (*BlocklistEnforcer, error) {
	wire.Build(
		server.NewProvider,
		core.NewCore,
		NewBlocklistEnforcer,
	)
	return nil, nil
}

func EnforceBlocklist() {
	color.Blue("Starting blocklist enforcement...")

	enforcer, err := InitializeBlocklistEnforcer()
	if err != nil {
		log.Fatalf("Failed to initialize blocklist enforcer: %v", err)
	}

	timeout := 30 * time.Minute
	if timeoutStr := os.Getenv("OPERATION_TIMEOUT_MINUTES"); timeoutStr != "" {
		if minutes, err := strconv.Atoi(timeoutStr); err == nil {
			timeout = time.Duration(minutes) * time.Minute
		}
	}

	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	if timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	if err := enforcer.Enforce(ctx); err != nil {
		log.Fatalf("Blocklist enforcement failed: %v", err)
	}

	color.Green("Blocklist enforcer stopped.")
}
