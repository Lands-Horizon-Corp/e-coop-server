package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
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

// executeLifecycle handles the fx application lifecycle
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
