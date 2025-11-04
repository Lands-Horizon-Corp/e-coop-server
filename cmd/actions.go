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
	v1 "github.com/Lands-Horizon-Corp/e-coop-server/server/controller"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/tokens"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/usecase"
	"github.com/fatih/color"
	"go.uber.org/fx"
)

func migrateDatabase() {
	color.Blue("Migrating database...")
	app := fx.New(
		fx.Provide(server.NewProvider, core.Newmodelcore),
		fx.Invoke(func(lc fx.Lifecycle, prov *server.Provider, mod *core.Core) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					if err := prov.Service.RunDatabase(ctx); err != nil {
						return err
					}
					if err := mod.Start(ctx); err != nil {
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
		fx.Provide(server.NewProvider, core.Newmodelcore, seeder.NewSeeder),
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
					if err := mod.Start(ctx); err != nil {
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
		fx.Provide(server.NewProvider, core.Newmodelcore, seeder.NewSeeder),
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
					if err := mod.Start(ctx); err != nil {
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
		fx.Provide(server.NewProvider, core.Newmodelcore),
		fx.Invoke(func(lc fx.Lifecycle, prov *server.Provider, mod *core.Core) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					if err := prov.Service.RunDatabase(ctx); err != nil {
						return err
					}
					if err := mod.Start(ctx); err != nil {
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
			core.Newmodelcore,
			v1.NewController,
			event.NewEvent,
			seeder.NewSeeder,
			tokens.NewUserToken,
			tokens.NewUserOrganizationToken,
			usecase.NewTransactionService,
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
					if err := mod.Start(ctx); err != nil {
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
		fx.Provide(server.NewProvider),
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
		fx.Provide(server.NewProvider, core.Newmodelcore, seeder.NewSeeder),
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
					if err := mod.Start(ctx); err != nil {
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
