package cmd

import (
	"strconv"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var command = &cobra.Command{
	Use:   "e-coop-server",
	Short: "E-Coop Server CLI - Financial cooperative management system",
	Long: `A comprehensive CLI tool for managing the E-Coop server application.
This tool provides commands for database management, cache operations, 
and server operations for your financial cooperative system.`,
	Run: func(_ *cobra.Command, _ []string) {
		helpers.PrintASCIIArt()
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(_ *cobra.Command, _ []string) {
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
				RunFunc: func(_ *cobra.Command, _ []string) error {
					return cleanCache()
				},
			},
		},
	},
	"security": {
		Parent: &cobra.Command{
			Use:   "security",
			Short: "Security utilities like HaGeZi blocklist",
		},
		Children: []CommandConfig{
			{
				Use:   "enforce",
				Short: "Update HaGeZi blocklist",
				RunFunc: func(_ *cobra.Command, _ []string) error {
					return enforceBlocklist()
				},
			},
			{
				Use:   "clear",
				Short: "Clear all blocked IPs from cache",
				RunFunc: func(_ *cobra.Command, _ []string) error {
					return clearBlockedIPs()
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
				RunFunc: func(_ *cobra.Command, _ []string) error {
					return migrateDatabase()
				},
			},
			{
				Use:   "seed",
				Short: "Seed the database with initial data",
				RunFunc: func(_ *cobra.Command, _ []string) error {
					return seedDatabase()
				},
			},
			{
				Use:   "performance-seed",
				Short: "Run database performance tests (creates test tables and data)",
				RunFunc: func(_ *cobra.Command, args []string) error {
					if len(args) == 0 {
						return seedDatabasePerformance(1)

					}
					multiplier, err := strconv.ParseInt(args[0], 10, 32)
					if err != nil {
						color.Red("Invalid multiplier, using default 1")
						return seedDatabasePerformance(1)

					}
					return seedDatabasePerformance(int32(multiplier))
				},
			},
			{
				Use:   "reset",
				Short: "Reset the database (drops and recreates)",
				RunFunc: func(_ *cobra.Command, _ []string) error {
					return resetDatabase()
				},
			},
			{
				Use:   "refresh",
				Short: "Reset the database and seed it with initial data",
				RunFunc: func(_ *cobra.Command, _ []string) error {
					return refreshDatabase()
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
		RunFunc: func(_ *cobra.Command, _ []string) error {
				if err := startServer(); err != nil {
					return err
				}
