package cmd

import (
	"github.com/spf13/cobra"
)

type CommandConfig struct {
	Use     string
	Short   string
	RunFunc func(cmd *cobra.Command, args []string) error
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
				RunE: func(cmd *cobra.Command, args []string) error {
					return cfg.RunFunc(cmd, args)
				},
			})
		}
	}
	for _, cfg := range standaloneCommands {
		command.AddCommand(&cobra.Command{
			Use:   cfg.Use,
			Short: cfg.Short,
			RunE: func(cmd *cobra.Command, args []string) error {
				return cfg.RunFunc(cmd, args)
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
