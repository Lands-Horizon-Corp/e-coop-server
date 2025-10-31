package cmd

import (
	"strconv"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var command = &cobra.Command{
	Use:   "e-coop-server",
	Short: "E-Coop Server CLI - Financial cooperative management system",
	Long: `A comprehensive CLI tool for managing the E-Coop server application.
This tool provides commands for database management, cache operations, 
and server operations for your financial cooperative system.`,
	Run: func(cmd *cobra.Command, args []string) {
		handlers.PrintASCIIArt()
		if err := cmd.Help(); err != nil {
			color.Red("Error displaying help: %v", err)
		}
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		color.Green("E-Coop Server v1.0.0")
	},
}

// Command groups with their configurations
var commandGroups = map[string]struct {
	Parent   *cobra.Command
	Children []CommandConfig
}{
	"cache": {
		Parent: &cobra.Command{
			Use:   "cache",
			Short: "Cache management commands",
		},
		Children: []CommandConfig{
			{
				Use:   "clean",
				Short: "Clean the application cache",
				RunFunc: func(cmd *cobra.Command, args []string) {
					cleanCache()
				},
			},
		},
	},
	"db": {
		Parent: &cobra.Command{
			Use:   "db",
			Short: "Database management commands",
		},
		Children: []CommandConfig{
			{
				Use:   "migrate",
				Short: "Automigrate all tables in the database",
				RunFunc: func(cmd *cobra.Command, args []string) {
					migrateDatabase()
				},
			},
			{
				Use:   "seed",
				Short: "Seed the database with initial data",
				RunFunc: func(cmd *cobra.Command, args []string) {
					seedDatabase()
				},
			},
			{
				Use:   "performance-seed",
				Short: "Run database performance tests (creates test tables and data)",
				RunFunc: func(cmd *cobra.Command, args []string) {
					multiplier, err := strconv.ParseInt(args[0], 2, 32)
					if err != nil {
						color.Red("Invalid multiplier, using default 1")
						multiplier = 1
					}
					seedDatabasePerformance(int32(multiplier))
				},
			},
			{
				Use:   "reset",
				Short: "Reset the database (drops and recreates)",
				RunFunc: func(cmd *cobra.Command, args []string) {
					resetDatabase()
				},
			},
			{
				Use:   "refresh",
				Short: "Reset the database and seed it with initial data",
				RunFunc: func(cmd *cobra.Command, args []string) {
					refreshDatabase()
				},
			},
		},
	},
}

// Standalone commands
var standaloneCommands = []CommandConfig{
	{
		Use:   "server",
		Short: "Start the main server",
		RunFunc: func(cmd *cobra.Command, args []string) {
			startServer()
		},
	},
}
