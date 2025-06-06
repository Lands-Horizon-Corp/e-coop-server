package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/lands-horizon/horizon-server/services/horizon"
)

func main() {
	if len(os.Args) < 2 {
		printWelcome()
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

func printWelcome() {
	horizon.PrintASCIIArt()
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
	// Add logic for cleaning cache
	color.Green("Cache cleaned successfully.")
}

func migrateDatabase() {
	color.Blue("Migrating database...")
	// Add logic for database migration (e.g., connecting to DB and running migrations)
	color.Green("Database migration completed successfully.")
}

func seedDatabase() {
	color.Blue("Seeding database...")
	// Add logic for seeding database
	color.Green("Database seeding completed successfully.")
}

func resetDatabase() {
	color.Blue("Resetting database...")
	// Add logic for resetting database (e.g., dropping and recreating tables)
	color.Green("Database reset completed successfully.")
}

func startServer() {
	color.Blue("Starting E-Coop Server on http://localhost:8000...")
	http.HandleFunc("/routes", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Available Routes:")
		fmt.Fprintln(w, "GET  /routes - View all available routes")
	})

	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		color.Red("Failed to start server: %v", err)
	}
}

// import (
// 	"context"
// 	"time"

// 	"github.com/lands-horizon/horizon-server/seeder"
// 	"github.com/lands-horizon/horizon-server/src"
// 	"github.com/lands-horizon/horizon-server/src/controller"
// 	"github.com/lands-horizon/horizon-server/src/cooperative_tokens"
// 	"github.com/lands-horizon/horizon-server/src/event"
// 	"github.com/lands-horizon/horizon-server/src/model"

// 	"go.uber.org/fx"
// )

// func main() {
// 	app := fx.New(
// 		// Set extended startup timeout
// 		fx.StartTimeout(2*time.Hour),

// 		// Provide application dependencies
// 		fx.Provide(
// 			src.NewProvider,
// 			src.NewValidator,
// 			model.NewModel,
// 			controller.NewController,
// 			event.NewEvent,
// 			seeder.NewSeeder,

// 			cooperative_tokens.NewUserToken,
// 			cooperative_tokens.NewTransactionBatchToken,
// 			cooperative_tokens.NewUserOrganizatonToken,
// 		),

// 		// Invoke the startup sequence
// 		fx.Invoke(func(
// 			lc fx.Lifecycle,
// 			ctrl *controller.Controller,
// 			mod *model.Model,
// 			prov *src.Provider,
// 			seed *seeder.Seeder,
// 		) error {
// 			// Register lifecycle hooks
// 			lc.Append(fx.Hook{
// 				OnStart: func(ctx context.Context) error {
// 					// Start each component in order
// 					if err := ctrl.Start(); err != nil {
// 						return err
// 					}
// 					if err := prov.Service.Run(ctx); err != nil {
// 						return err
// 					}
// 					if err := mod.Start(); err != nil {
// 						return err
// 					}
// 					if err := seed.Run(ctx); err != nil {
// 						return err
// 					}
// 					return nil
// 				},

// 				OnStop: func(ctx context.Context) error {
// 					// Gracefully stop the service
// 					return prov.Service.Stop(ctx)
// 				},
// 			})

// 			return nil
// 		}),
// 	)

// 	app.Run()
// }
