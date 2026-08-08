package main

import "github.com/spf13/cobra"

var version = "dev"

func newRootCommand() *cobra.Command {
	command := &cobra.Command{
		Use:          "watt",
		Short:        "Local pipeline execution and verification engine",
		SilenceUsage: true,
		Version:      version,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
	command.SetVersionTemplate("watt {{.Version}}\n")
	command.AddCommand(
		&cobra.Command{Use: "run", Run: func(*cobra.Command, []string) {}},
		&cobra.Command{Use: "check", Run: func(*cobra.Command, []string) {}},
	)

	return command
}
