package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/bext1998/WattCIAutomationEngine/internal/pipeline"
)

var version = "dev"

func wrapUsageError(err error) error {
	return &exitError{code: EXIT_USAGE, err: err}
}

func usageArgs(command *cobra.Command, args []string) error {
	if err := cobra.NoArgs(command, args); err != nil {
		return wrapUsageError(err)
	}
	return nil
}

func newRootCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "watt",
		Short: "Local pipeline execution and verification engine",
		Args:  usageArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			return command.Help()
		},
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       version,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
	command.SetVersionTemplate("watt {{.Version}}\n")
	command.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return wrapUsageError(err)
	})
	notImplemented := func(command *cobra.Command, _ []string) error {
		return &exitError{
			code: EXIT_INTERNAL_ERROR,
			err:  fmt.Errorf("%s is not implemented", command.Name()),
		}
	}
	check := &cobra.Command{
		Use:  "check",
		Args: usageArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			repoRoot, err := os.Getwd()
			if err != nil {
				return &exitError{code: EXIT_INTERNAL_ERROR, err: fmt.Errorf("get repository root: %w", err)}
			}

			definition, err := pipeline.Load(filepath.Join(repoRoot, "watt.yaml"))
			if err != nil {
				return &exitError{code: EXIT_INVALID_PIPELINE, err: err}
			}
			if err := definition.Validate(); err != nil {
				return &exitError{code: EXIT_INVALID_PIPELINE, err: err}
			}
			return nil
		},
	}
	command.AddCommand(
		&cobra.Command{Use: "run", Args: usageArgs, RunE: notImplemented},
		check,
	)

	return command
}
