package cmd

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// CommandConfig defines a command configuration
type CommandConfig struct {
	Use     string
	Short   string
	RunFunc func(cmd *cobra.Command, args []string)
}

func init() {
	command.AddCommand(versionCmd)

	// Register command groups (parent commands with subcommands)
	for _, group := range commandGroups {
		command.AddCommand(group.Parent)
		for _, childConfig := range group.Children {
			childCmd := &cobra.Command{
				Use:   childConfig.Use,
				Short: childConfig.Short,
				Run:   childConfig.RunFunc,
			}
			group.Parent.AddCommand(childCmd)
		}
	}

	// Register standalone commands
	for _, cmdConfig := range standaloneCommands {
		cmd := &cobra.Command{
			Use:   cmdConfig.Use,
			Short: cmdConfig.Short,
			Run:   cmdConfig.RunFunc,
		}
		command.AddCommand(cmd)
	}
}

// Execute runs the root CLI command and exits on failure.
// It is intended to be called from main.main to start the CLI.
func Execute() {
	if err := command.Execute(); err != nil {
		panic(err)
	}
}

// getOperationTimeout returns timeout from environment or default
func getOperationTimeout(defaultTimeout time.Duration) time.Duration {
	if timeoutStr := os.Getenv("OPERATION_TIMEOUT_MINUTES"); timeoutStr != "" {
		if minutes, err := strconv.Atoi(timeoutStr); err == nil {
			return time.Duration(minutes) * time.Minute
		}
	}
	return defaultTimeout
}

// executeLifecycle handles the fx application lifecycle
func executeLifecycle(app *fx.App) {
	timeout := getOperationTimeout(4 * time.Hour)
	executeLifecycleWithTimeout(app, timeout)
}

// executeLifecycleWithTimeout handles the fx application lifecycle with custom timeout
func executeLifecycleWithTimeout(app *fx.App, timeout time.Duration) {
	startCtx, startCancel := context.WithTimeout(context.Background(), timeout)
	if err := app.Start(startCtx); err != nil {
		// Ensure we cancel the start context before logging a fatal error so
		// any cancel cleanup runs as expected.
		startCancel()
		log.Fatalf("Failed to start: %v", err)
	}
	// Clean up start context resources after successful start.
	startCancel()

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Minute)
	if err := app.Stop(stopCtx); err != nil {
		stopCancel()
		log.Fatalf("Failed to stop: %v", err)
	}
	// Clean up stop context resources after successful stop.
	stopCancel()
}
