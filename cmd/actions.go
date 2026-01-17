package cmd

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	v1 "github.com/Lands-Horizon-Corp/e-coop-server/src/controller/v1"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/fatih/color"
)

func enforceBlocklist() error {
	return horizon.WithHorizonService(
		horizon.DefaultHorizonRunnerParams{
			TimeoutValue:       30 * time.Minute,
			OnStartMessageText: "Enforcing HaGeZi blocklist...",
			OnStopMessageText:  "Blocklist enforcement stopped",
			HandlerFunc: func(ctx context.Context, service *horizon.HorizonService) error {
				return service.Security.Firewall(ctx, func(ip, host string) {
					cacheKey := "blocked_ip:" + ip
					timestamp := float64(time.Now().Unix())

					if err := service.Cache.ZAdd(ctx, "blocked_ips_registry", timestamp, ip); err != nil {
						color.Red("Failed to add IP %s: %v", ip, err)
					}
					if err := service.Cache.Set(ctx, cacheKey, host, 60*24*time.Hour); err != nil {
						color.Red("Failed to cache IP %s: %v", ip, err)
					}
					color.Yellow("Cached blocked IP %s from host %s", ip, host)
				})
			},
		},
	)
}

func clearBlockedIPs() error {
	return horizon.WithHorizonService(
		horizon.DefaultHorizonRunnerParams{
			TimeoutValue:       30 * time.Minute,
			OnStartMessageText: "Clearing blocked IPs from cache...",
			OnStopMessageText:  "Blocked IPs cleared successfully.",
			HandlerFunc: func(ctx context.Context, service *horizon.HorizonService) error {
				keys, err := service.Cache.Keys(ctx, "blocked_ip:*")
				if err != nil {
					color.Red("Failed to get blocked IP keys: %v", err)
					return err
				}
				count := 0
				for _, key := range keys {
					if err := service.Cache.Delete(ctx, key); err != nil {
						color.Red("Failed to delete key %s: %v", key, err)
					} else {
						count++
					}
				}
				color.Green("Cleared %d blocked IP entries from cache", count)
				return nil
			},
		},
	)
}

func migrateDatabase() error {
	return horizon.WithHorizonService(
		horizon.DefaultHorizonRunnerParams{
			TimeoutValue:       30 * time.Minute,
			OnStartMessageText: "Migrating database...",
			OnStopMessageText:  "Database migration completed.",
			HandlerFunc: func(ctx context.Context, service *horizon.HorizonService) error {

				if err := service.Database.Client().AutoMigrate(types.Models()...); err != nil {
					return err
				}
				return nil
			},
		},
	)
}

func seedDatabase() error {
	return horizon.WithHorizonService(
		horizon.DefaultHorizonRunnerParams{
			TimeoutValue:       2 * time.Hour,
			OnStartMessageText: "Seeding database...",
			OnStopMessageText:  "Database seeding completed.",
			HandlerFunc: func(ctx context.Context, service *horizon.HorizonService) error {
				if err := core.Seed(ctx, service, 5); err != nil {
					return err
				}
				return nil
			},
		},
	)
}

func seedDatabasePerformance(multiplier int32) error {
	return horizon.WithHorizonService(
		horizon.DefaultHorizonRunnerParams{
			TimeoutValue:       2 * time.Hour,
			OnStartMessageText: "Seeding database... with performance seed",
			OnStopMessageText:  "Database seeding completed.",
			HandlerFunc: func(ctx context.Context, service *horizon.HorizonService) error {
				if err := core.Seed(ctx, service, multiplier); err != nil {
					return err
				}
				return nil
			},
		},
	)
}

func resetDatabase() error {
	return horizon.WithHorizonService(
		horizon.DefaultHorizonRunnerParams{
			TimeoutValue:       30 * time.Minute,
			OnStartMessageText: "Resetting database...",
			OnStopMessageText:  "Database reset completed successfully.",
			HandlerFunc: func(ctx context.Context, service *horizon.HorizonService) error {
				if err := service.Storage.RemoveAllFiles(ctx); err != nil {
					return err
				}
				models := types.Models()
				if err := service.Database.Client().Migrator().DropTable(models...); err != nil {
					return err
				}
				if err := service.Database.Client().AutoMigrate(models...); err != nil {
					return err
				}
				return nil
			},
		},
	)
}

func cleanCache() error {
	return horizon.WithHorizonService(
		horizon.DefaultHorizonRunnerParams{
			TimeoutValue:       30 * time.Minute,
			OnStartMessageText: "Cleaning cache...",
			OnStopMessageText:  "Cache cleaned successfully.",
			HandlerFunc: func(ctx context.Context, service *horizon.HorizonService) error {

				if err := service.Cache.Flush(ctx); err != nil {
					return err
				}
				return nil
			},
		},
	)
}

func refreshDatabase() error {
	return horizon.WithHorizonService(
		horizon.DefaultHorizonRunnerParams{
			TimeoutValue:       2 * time.Hour,
			OnStartMessageText: "Refreshing database...",
			OnStopMessageText:  "Database refreshed successfully.",
			HandlerFunc: func(ctx context.Context, service *horizon.HorizonService) error {
				models := types.Models()
				if err := service.Cache.Flush(ctx); err != nil {
					return err
				}
				if err := service.Storage.RemoveAllFiles(ctx); err != nil {
					return err
				}
				if err := service.Database.Client().Migrator().DropTable(models...); err != nil {
					return err
				}
				if err := service.Database.Client().AutoMigrate(models...); err != nil {
					return err
				}
				if err := core.Seed(ctx, service, 5); err != nil {
					return err
				}
				return nil
			},
		},
	)
}

func startServer() error {
	forceLifeTime := true
	return horizon.WithHorizonService(
		horizon.DefaultHorizonRunnerParams{
			ForceLifetimeFunc:  &forceLifeTime,
			TimeoutValue:       5 * time.Minute,
			OnStartMessageText: "Starting Server ...",
			OnStopMessageText:  "Server started successfully.",
			HandlerFunc: func(ctx context.Context, service *horizon.HorizonService) error {
				if err := v1.Controllers(service); err != nil {
					return err
				}
				if err := service.RunLifeTime(ctx); err != nil {
					return err
				}
				return nil
			},
		},
	)
}
