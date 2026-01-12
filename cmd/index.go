package cmd

import (
	"github.com/spf13/cobra"
)

type CommandConfig struct {
	Use     string
	Short   string
	RunFunc func(cmd *cobra.Command, args []string)
}

func buildCommands() {
	command.AddCommand(versionCmd)

	for _, group := range commandGroups {
		command.AddCommand(group.Parent)

		for _, childConfig := range group.Children {
			cfg := childConfig
			group.Parent.AddCommand(&cobra.Command{
				Use:   cfg.Use,
				Short: cfg.Short,
				Run: func(cmd *cobra.Command, args []string) {
					cfg.RunFunc(cmd, args)
				},
			})
		}
	}

	for _, cfg := range standaloneCommands {
		cfg := cfg
		command.AddCommand(&cobra.Command{
			Use:   cfg.Use,
			Short: cfg.Short,
			Run: func(cmd *cobra.Command, args []string) {
				cfg.RunFunc(cmd, args)
			},
		})
	}
}

func Execute() {
	buildCommands()
	if err := command.Execute(); err != nil {
		panic(err)
	}
}

// func executeLifecycle(app *fx.App) {
// 	timeout := 4 * time.Hour
// 	if timeoutStr := os.Getenv("OPERATION_TIMEOUT_MINUTES"); timeoutStr != "" {
// 		if minutes, err := strconv.Atoi(timeoutStr); err == nil {
// 			timeout = time.Duration(minutes) * time.Minute
// 		}
// 	}
// 	executeLifecycleWithTimeout(app, timeout)
// }

// func executeLifecycleWithTimeout(app *fx.App, timeout time.Duration) {
// 	startCtx, startCancel := context.WithTimeout(context.Background(), timeout)
// 	if err := app.Start(startCtx); err != nil {
// 		startCancel()
// 		log.Fatalf("Failed to start: %v", err)
// 	}
// 	startCancel()
// 	stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Minute)
// 	if err := app.Stop(stopCtx); err != nil {
// 		stopCancel()
// 		log.Fatalf("Failed to stop: %v", err)
// 	}
// 	stopCancel()
// }
