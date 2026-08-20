package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bext1998/WattCIAutomationEngine/internal/env"
	"github.com/bext1998/WattCIAutomationEngine/internal/orchestrator"
	"github.com/bext1998/WattCIAutomationEngine/internal/pipeline"
	"github.com/bext1998/WattCIAutomationEngine/internal/result"
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

func runArgs(_ *cobra.Command, args []string) error {
	if len(args) > 1 {
		return wrapUsageError(fmt.Errorf("run accepts at most one pipeline name, got %d arguments", len(args)))
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
			if err := writeBrand(command.OutOrStdout()); err != nil {
				return &exitError{code: EXIT_INTERNAL_ERROR, err: err}
			}
			command.Short = ""
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
	var output string
	run := &cobra.Command{
		Use:  "run [pipeline]",
		Args: runArgs,
		RunE: func(command *cobra.Command, args []string) error {
			if output != "" && output != "json" {
				return wrapUsageError(fmt.Errorf("unsupported --output value %q; only \"json\" is supported", output))
			}

			repoRoot, err := os.Getwd()
			if err != nil {
				return &exitError{code: EXIT_INTERNAL_ERROR, err: fmt.Errorf("get repository root: %w", err)}
			}

			pipelineName := ""
			if len(args) == 1 {
				pipelineName = args[0]
			}

			ctx, stop := signal.NotifyContext(command.Context(), os.Interrupt)
			defer stop()

			stdout := command.OutOrStdout()
			stderr := command.ErrOrStderr()
			stepStdout := stdout
			if output == "json" {
				stepStdout = stderr
			}
			outcome, err := orchestrator.Run(orchestrator.Options{
				RepoRoot:     repoRoot,
				PipelineName: pipelineName,
				Version:      version,
				Stdout:       stepStdout,
				Stderr:       stderr,
				Context:      ctx,
			})
			if output == "json" && outcome.Assembled {
				contents, marshalErr := result.Marshal(outcome.Result)
				if marshalErr != nil {
					return &exitError{code: EXIT_INTERNAL_ERROR, err: marshalErr}
				}
				if _, writeErr := stdout.Write(contents); writeErr != nil {
					return &exitError{code: EXIT_INTERNAL_ERROR, err: fmt.Errorf("write JSON result to stdout: %w", writeErr)}
				}
			}
			if err != nil {
				return &exitError{code: int(outcome.Code), err: err}
			}
			if outcome.Code != orchestrator.ExitSuccess {
				return &exitError{code: int(outcome.Code), err: fmt.Errorf("pipeline finished with exit code %d", outcome.Code)}
			}
			return nil
		},
	}
	run.Flags().StringVar(&output, "output", "", "write the final result to stdout (json)")
	command.AddCommand(
		run,
		check,
	)

	return command
}
