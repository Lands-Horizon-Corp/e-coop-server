package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/lands-horizon/horizon-server/seeder"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src"
	"github.com/lands-horizon/horizon-server/src/controller"
	"github.com/lands-horizon/horizon-server/src/cooperative_tokens"
	"github.com/lands-horizon/horizon-server/src/event"
	"github.com/lands-horizon/horizon-server/src/model"
	"go.uber.org/fx"
)

func main() {
	if len(os.Args) < 2 {
		horizon.PrintASCIIArt()
		printHelp()
		return
	}

	command := strings.ToLower(os.Args[1])
	switch command {
	case "cache:clean":
		cleanCache()
	case "db:migrate":
		migrateDatabase()
	case "db:seed":
		seedDatabase()
	case "db:reset":
		resetDatabase()
	case "server":
		startServer()
	default:
		color.Red("Unknown command: %s\n", command)
		printHelp()
	}
}

func printHelp() {
	color.Yellow("Usage:")
	color.Green("  go run main.go <command>")
	fmt.Println()
	color.Yellow("Available Commands:")
	color.Green("  cache:clean      ")
	fmt.Println("      Clean the application cache")
	color.Green("  db:migrate       ")
	fmt.Println("      Automigrate all tables in the database")
	color.Green("  db:seed          ")
	fmt.Println("      Seed the database with initial data")
	color.Green("  db:reset         ")
	fmt.Println("      Reset the database (drops and recreates)")
	color.Green("  server           ")
	fmt.Println("      Start the main server")
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

func migrateDatabase() {
	color.Blue("Migrating database...")
	app := fx.New(
		fx.Provide(src.NewProvider, model.NewModel),
		fx.Invoke(func(lc fx.Lifecycle, prov *src.Provider, mod *model.Model) {
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
		fx.Provide(src.NewProvider, model.NewModel, seeder.NewSeeder),
		fx.Invoke(func(lc fx.Lifecycle, prov *src.Provider, mod *model.Model, seed *seeder.Seeder) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					if err := prov.Service.RunDatabase(ctx); err != nil {
						return err
					}
					if err := prov.Service.RunStorage(ctx); err != nil {
						return err
					}
					if err := mod.Start(ctx); err != nil {
						return err
					}
					if err := seed.Run(ctx); err != nil {
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
		fx.Provide(src.NewProvider, model.NewModel),
		fx.Invoke(func(lc fx.Lifecycle, prov *src.Provider, mod *model.Model) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					if err := prov.Service.RunDatabase(ctx); err != nil {
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
			model.NewModel,
			controller.NewController,
			event.NewEvent,
			seeder.NewSeeder,
			cooperative_tokens.NewUserToken,
			cooperative_tokens.NewTransactionBatchToken,
			cooperative_tokens.NewUserOrganizatonToken,
		),
		fx.Invoke(func(lc fx.Lifecycle, ctrl *controller.Controller, mod *model.Model, prov *src.Provider) error {
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

func executeLifecycle(app *fx.App) {
	startCtx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	if err := app.Start(startCtx); err != nil {
		log.Fatalf("Failed to start: %v", err)
	}

	stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	if err := app.Stop(stopCtx); err != nil {
		log.Fatalf("Failed to stop: %v", err)
	}
}
