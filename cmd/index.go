package cmd

import (
	"context"
	"fmt"
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

func Execute() {
	if err := command.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
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
	startCtx, cancel := context.WithTimeout(context.Background(), timeout)
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
