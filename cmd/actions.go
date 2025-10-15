package cmd

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/seeder"
	"github.com/Lands-Horizon-Corp/e-coop-server/src"
	controller_v1 "github.com/Lands-Horizon-Corp/e-coop-server/src/controller/v1"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/cooperative_tokens"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/event"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model/model_core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/service"
	"github.com/fatih/color"
	"go.uber.org/fx"
)

func migrateDatabase() {
	color.Blue("Migrating database...")
	app := fx.New(
		fx.Provide(src.NewProvider, model_core.NewModelCore),
		fx.Invoke(func(lc fx.Lifecycle, prov *src.Provider, mod *model_core.ModelCore) {
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
		fx.Provide(src.NewProvider, model_core.NewModelCore, seeder.NewSeeder),
		fx.Invoke(func(
			lc fx.Lifecycle,
			prov *src.Provider,
			mod *model_core.ModelCore,
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

	executeLifecycle(app)
	color.Green("Database seeding completed successfully.")
}

func seedDatabasePerformance(multiplier int32) {
	color.Blue("Seeding database...")
	app := fx.New(
		fx.Provide(src.NewProvider, model_core.NewModelCore, seeder.NewSeeder),
		fx.Invoke(func(
			lc fx.Lifecycle,
			prov *src.Provider,
			mod *model_core.ModelCore,
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
		fx.Provide(src.NewProvider, model_core.NewModelCore),
		fx.Invoke(func(lc fx.Lifecycle, prov *src.Provider, mod *model_core.ModelCore) {
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
			src.NewProvider,
			src.NewValidator,
			model_core.NewModelCore,
			controller_v1.NewController,
			event.NewEvent,
			seeder.NewSeeder,
			cooperative_tokens.NewUserToken,
			cooperative_tokens.NewUserOrganizationToken,
			service.NewTransactionService,
		),
		fx.Invoke(func(lc fx.Lifecycle, ctrl *controller_v1.Controller, mod *model_core.ModelCore, prov *src.Provider) error {
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
		fx.Provide(src.NewProvider),
		fx.Invoke(func(lc fx.Lifecycle, prov *src.Provider) {
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
		fx.Provide(src.NewProvider, model_core.NewModelCore, seeder.NewSeeder),
		fx.Invoke(func(lc fx.Lifecycle, prov *src.Provider, mod *model_core.ModelCore, seed *seeder.Seeder) {
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

	executeLifecycle(app)
	color.Green("Database reset completed successfully.")
}
