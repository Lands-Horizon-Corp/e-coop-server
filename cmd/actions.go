// Package cmd contains CLI command wiring and helper actions for the e-coop-server.
//
// It defines CLI commands used for database migration, seeding, cache management,
// and running the server in development or test environments.
package cmd

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/seeder"
	"github.com/Lands-Horizon-Corp/e-coop-server/server"
	v1 "github.com/Lands-Horizon-Corp/e-coop-server/server/controller/v1"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/report"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/tokens"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/usecase"
	"github.com/fatih/color"
	"go.uber.org/fx"
)

func enforceBlocklist() {
	color.Blue("Enforcing HaGeZi blocklist...")
	app := fx.New(
		fx.Provide(
			server.NewProvider,
			core.NewCore,
		),
		fx.Invoke(func(lc fx.Lifecycle, prov *server.Provider, mod *core.Core) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					if err := prov.Service.RunCache(ctx); err != nil {
						return err
					}
					if err := prov.Service.Security.Firewall(ctx, func(ip, host string) {
						cacheKey := "blocked_ip:" + ip
						timestamp := float64(time.Now().Unix())

						// Use ZAdd for consistent sorted set tracking
						if err := prov.Service.Cache.ZAdd(ctx, "blocked_ips_registry", timestamp, ip); err != nil {
							color.Red("Failed to add IP %s to registry: %v", ip, err)
						}

						// Keep the original Set for backward compatibility with middleware
						if err := prov.Service.Cache.Set(ctx, cacheKey, host, 60*24*time.Hour); err != nil {
							color.Red("Failed to cache IP %s: %v", ip, err)
						}
						color.Yellow("Cached blocked IP %s from host %s", ip, host)
					}); err != nil {
						return err
					}
					return nil
				},
			})
		}),
	)
	executeLifecycle(app)
	color.Green("HaGeZi blocklist enforced and cached successfully.")
}

// clearBlockedIPs removes all blocked IP entries from Redis cache
func clearBlockedIPs() {
	color.Blue("Clearing blocked IPs from cache...")
	app := fx.New(
		fx.Provide(
			server.NewProvider,
		),
		fx.Invoke(func(lc fx.Lifecycle, prov *server.Provider) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					if err := prov.Service.RunCache(ctx); err != nil {
						return err
					}

					// Get all blocked IP keys
					keys, err := prov.Service.Cache.Keys(ctx, "blocked_ip:*")
					if err != nil {
						color.Red("Failed to get blocked IP keys: %v", err)
						return err
					}

					// Delete each blocked IP key
					count := 0
					for _, key := range keys {
						if err := prov.Service.Cache.Delete(ctx, key); err != nil {
							color.Red("Failed to delete key %s: %v", key, err)
						} else {
							count++
						}
					}

					color.Green("Cleared %d blocked IP entries from cache", count)
					return nil
				},
			})
		}),
	)
	executeLifecycle(app)
	color.Green("Blocked IPs cleared successfully.")
}

func migrateDatabase() {
	color.Blue("Migrating database...")
	app := fx.New(
		fx.Provide(
			server.NewProvider,
			core.NewCore,
		),
		fx.Invoke(func(lc fx.Lifecycle, prov *server.Provider, mod *core.Core) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					if err := prov.Service.RunDatabase(ctx); err != nil {
						return err
					}
					if err := mod.Start(); err != nil {
						return err
					}
					if err := prov.Service.Database.Client().AutoMigrate(mod.Migration...); err != nil {
						return err
					}
					return nil
				},
			})
		}),
	)
	executeLifecycle(app)
	color.Green("Database migration completed successfully.")
}

func seedDatabase() {
	color.Blue("Seeding database...")
	app := fx.New(
		fx.StartTimeout(3*time.Hour), // Longer timeout for seeding
		fx.Provide(
			server.NewProvider,
			core.NewCore,
			seeder.NewSeeder,
		),
		fx.Invoke(func(
			lc fx.Lifecycle,
			prov *server.Provider,
			mod *core.Core,
			seed *seeder.Seeder,
		) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					if err := prov.Service.RunDatabase(ctx); err != nil {
						return err
					}
					if err := prov.Service.RunStorage(ctx); err != nil {
						return err
					}
					if err := prov.Service.RunStorage(ctx); err != nil {
						return err
					}
					if err := prov.Service.RunBroker(ctx); err != nil {
						return err
					}
					if err := mod.Start(); err != nil {
						return err
					}
					if err := seed.Run(ctx, 5); err != nil {
						return err
					}
					return nil
				},
			})
		}),
	)

	executeLifecycleWithTimeout(app, 3*time.Hour)
	color.Green("Database seeding completed successfully.")
}

func seedDatabasePerformance(multiplier int32) {
	color.Blue("Seeding database...")
	app := fx.New(
		fx.Provide(
			server.NewProvider,
			core.NewCore,
			seeder.NewSeeder,
		),
		fx.Invoke(func(
			lc fx.Lifecycle,
			prov *server.Provider,
			mod *core.Core,
			seed *seeder.Seeder,
		) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					if err := prov.Service.RunDatabase(ctx); err != nil {
						return err
					}
					if err := prov.Service.RunStorage(ctx); err != nil {
						return err
					}
					if err := prov.Service.RunStorage(ctx); err != nil {
						return err
					}
					if err := prov.Service.RunBroker(ctx); err != nil {
						return err
					}
					if err := mod.Start(); err != nil {
						return err
					}
					if err := seed.Run(ctx, multiplier); err != nil {
						return err
					}
					return nil
				},
			})
		}),
	)

	executeLifecycle(app)
	color.Green("Database seeding completed successfully.")
}

func resetDatabase() {
	color.Blue("Resetting database...")
	app := fx.New(
		fx.Provide(
			server.NewProvider,
			core.NewCore,
		),
		fx.Invoke(func(lc fx.Lifecycle, prov *server.Provider, mod *core.Core) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					if err := prov.Service.RunDatabase(ctx); err != nil {
						return err
					}
					if err := mod.Start(); err != nil {
						return err
					}
					if err := prov.Service.RunStorage(ctx); err != nil {
						return err
					}
					if err := prov.Service.Storage.RemoveAllFiles(ctx); err != nil {
						return err
					}
					if err := prov.Service.Database.Client().Migrator().DropTable(mod.Migration...); err != nil {
						return err
					}
					if err := prov.Service.Database.Client().AutoMigrate(mod.Migration...); err != nil {
						return err
					}
					return nil
				},
			})
		}),
	)

	executeLifecycle(app)
	color.Green("Database reset completed successfully.")
}

func startServer() {
	app := fx.New(
		fx.StartTimeout(2*time.Hour),
		fx.Provide(
			server.NewProvider,
			server.NewValidator,
			core.NewCore,
			v1.NewController,
			event.NewEvent,
			report.NewReports,
			seeder.NewSeeder,
			tokens.NewUserToken,
			tokens.NewUserOrganizationToken,
			usecase.NewUsecaseService,
		),
		fx.Invoke(func(lc fx.Lifecycle, ctrl *v1.Controller, mod *core.Core, prov *server.Provider) error {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					if err := ctrl.Start(); err != nil {
						return err
					}
					if err := prov.Service.Run(ctx); err != nil {
						return err
					}
					if err := mod.Start(); err != nil {
						return err
					}
					return nil
				},
				OnStop: func(ctx context.Context) error {
					return prov.Service.Stop(ctx)
				},
			})
			return nil
		}),
	)

	app.Run()
}

func cleanCache() {
	color.Blue("Cleaning cache...")
	app := fx.New(
		fx.Provide(
			server.NewProvider,
		),
		fx.Invoke(func(lc fx.Lifecycle, prov *server.Provider) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					if err := prov.Service.RunCache(ctx); err != nil {
						return err
					}
					if err := prov.Service.Cache.Flush(ctx); err != nil {
						return err
					}
					return nil
				},
			})
		}),
	)

	executeLifecycle(app)
	color.Green("Cache cleaned successfully.")
}

func refreshDatabase() {
	color.Blue("Refreshing database...")
	app := fx.New(
		fx.StartTimeout(3*time.Hour), // Longer timeout for full refresh
		fx.Provide(
			server.NewProvider,
			core.NewCore,
			seeder.NewSeeder,
			report.NewReports),
		fx.Invoke(func(lc fx.Lifecycle, prov *server.Provider, mod *core.Core, seed *seeder.Seeder) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					if err := prov.Service.RunDatabase(ctx); err != nil {
						return err
					}
					if err := prov.Service.RunStorage(ctx); err != nil {
						return err
					}
					if err := prov.Service.Storage.RemoveAllFiles(ctx); err != nil {
						return err
					}
					if err := prov.Service.RunBroker(ctx); err != nil {
						return err
					}
					if err := mod.Start(); err != nil {
						return err
					}
					if err := prov.Service.Database.Client().Migrator().DropTable(mod.Migration...); err != nil {
						return err
					}
					if err := prov.Service.Database.Client().AutoMigrate(mod.Migration...); err != nil {
						return err
					}
					if err := seed.Run(ctx, 5); err != nil {
						return err
					}
					return nil
				},
			})
		}),
	)

	executeLifecycleWithTimeout(app, 3*time.Hour)
	color.Green("Database reset completed successfully.")
}
