package cmd

import (
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/ui"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func Execute() {
	command := &cobra.Command{
		Use:   "e-coop-server",
		Short: "E-Coop Server CLI - Financial cooperative management system",
		Long: `A comprehensive CLI tool for managing the E-Coop server application.
This tool provides commands for database management, cache operations, 
and server operations for your financial cooperative system.`,
		Run: func(_ *cobra.Command, _ []string) {
			ui.Logo()
		},
	}
	command.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(_ *cobra.Command, _ []string) {
			color.Green("E-Coop Server v1.0.0")
		},
	})
	for _, cfg := range Register() {
		cfgCopy := cfg
		command.AddCommand(&cobra.Command{
			Use:   cfgCopy.Use,
			Short: cfgCopy.Short,
			RunE: func(cmd *cobra.Command, args []string) error {
				return cfgCopy.RunFunc(cmd, args)
			},
		})
	}
	if err := command.Execute(); err != nil {
		panic(err)
	}
}
