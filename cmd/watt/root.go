package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bext1998/WattCIAutomationEngine/internal/env"
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

type environmentRequirement struct {
	kind string
	name string
}

func probeEnvironment(definition pipeline.PipelineFile) error {
	pipelineNames := make([]string, 0, len(definition.Pipelines))
	for name := range definition.Pipelines {
		pipelineNames = append(pipelineNames, name)
	}
	sort.Strings(pipelineNames)

	missing := make([]environmentRequirement, 0)
	seen := make(map[environmentRequirement]struct{})
	for _, pipelineName := range pipelineNames {
		for _, step := range definition.Pipelines[pipelineName].Steps {
			requirement := environmentRequirement{}
			if strings.TrimSpace(step.Exec) != "" {
				requirement = environmentRequirement{kind: "command", name: step.Exec}
			} else {
				requirement = environmentRequirement{kind: "shell", name: step.Shell}
			}

			if _, exists := seen[requirement]; exists {
				continue
			}
			seen[requirement] = struct{}{}

			if _, err := env.ResolveExecutable(requirement.name); err != nil {
				missing = append(missing, requirement)
			}
		}
	}

	if len(missing) == 0 {
		return nil
	}

	items := make([]string, 0, len(missing))
	for _, requirement := range missing {
		items = append(items, fmt.Sprintf("missing %s %q", requirement.kind, requirement.name))
	}
	return fmt.Errorf("environment unavailable: %s", strings.Join(items, "; "))
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
	var checkEnv bool
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
			if checkEnv {
				if err := probeEnvironment(definition); err != nil {
					return &exitError{code: EXIT_ENVIRONMENT_UNAVAILABLE, err: err}
				}
			}
			return nil
		},
	}
	check.Flags().BoolVar(&checkEnv, "env", false, "check required commands and shells in PATH")
	command.AddCommand(
		&cobra.Command{Use: "run", Args: usageArgs, RunE: notImplemented},
		check,
	)

	return command
}
