package cable

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server"
	"github.com/fatih/color"
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

func CleanCache() {
	color.Blue("Cleaning cache...")

	cleaner, err := InitializeCacheCleaner()
	if err != nil {
		log.Fatalf("Failed to initialize cache cleaner: %v", err)
	}

	timeout := 10 * time.Minute
	if timeoutStr := os.Getenv("OPERATION_TIMEOUT_MINUTES"); timeoutStr != "" {
		if minutes, err := strconv.Atoi(timeoutStr); err == nil {
			timeout = time.Duration(minutes) * time.Minute
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := cleaner.Clean(ctx); err != nil {
		log.Fatalf("Cache cleaning failed: %v", err)
	}

	color.Green("Cache cleaned successfully.")
}
