package orchestrator

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/bext1998/WattCIAutomationEngine/internal/pipeline"
	"github.com/bext1998/WattCIAutomationEngine/internal/result"
	"github.com/bext1998/WattCIAutomationEngine/internal/runner"
)

type ExitCode int

const (
	ExitSuccess                ExitCode = 0
	ExitStepFailed             ExitCode = 1
	ExitInvalidPipeline        ExitCode = 2
	ExitEnvironmentUnavailable ExitCode = 3
	ExitCancelled              ExitCode = 4
	ExitInternalError          ExitCode = 5
)

type Options struct {
	RepoRoot     string
	PipelinePath string
	PipelineName string
	Stdout       io.Writer
	Stderr       io.Writer
	Context      context.Context

	// ConfirmDeadline is forwarded to runner.Step.ConfirmDeadline (and from
	// there to proc.Spec.ConfirmDeadline). A non-positive value uses the default
	// (5 seconds, A-10). Production callers must leave it unset; tests may
	// shrink it.
	ConfirmDeadline time.Duration
}

type Outcome struct {
	Code   ExitCode
	Result result.Result
}

func Run(options Options) (Outcome, error) {
	ctx := options.Context
	if ctx == nil {
		ctx = context.Background()
	}

	// Cancellation before load/validate: no result.json may be written (§4.2).
	if ctx.Err() != nil {
		return Outcome{Code: ExitCancelled}, nil
	}

	repoRoot := options.RepoRoot
	if repoRoot == "" {
		var err error
		repoRoot, err = os.Getwd()
		if err != nil {
			return Outcome{Code: ExitInternalError}, fmt.Errorf("get repository root: %w", err)
		}
	}
	pipelinePath := options.PipelinePath
	if pipelinePath == "" {
		pipelinePath = filepath.Join(repoRoot, "watt.yaml")
	}
	stdout := options.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := options.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	definition, err := pipeline.Load(pipelinePath)
	if err != nil {
		return Outcome{Code: ExitInvalidPipeline}, err
	}
	if err := definition.Validate(); err != nil {
		return Outcome{Code: ExitInvalidPipeline}, err
	}

	pipelineName := options.PipelineName
	if pipelineName == "" {
		pipelineName = "default"
	}
	selected, exists := definition.Pipelines[pipelineName]
	if !exists {
		return Outcome{Code: ExitInvalidPipeline}, missingPipelineError(pipelineName, definition.Pipelines)
	}

	startedAt := time.Now()
	final := result.Result{
		SchemaVersion: result.SchemaVersion,
		Pipeline:      pipelineName,
		Status:        "success",
		Steps:         make([]result.Step, 0, len(selected.Steps)),
		StartedAt:     startedAt.Format(time.RFC3339Nano),
		WattVersion:   "dev",
		Environment: result.Environment{
			OS:   runtime.GOOS,
			Arch: runtime.GOARCH,
		},
	}

	// Cancellation after load+validate but before any step: write an empty
	// cancelled result (§4.2).
	if ctx.Err() != nil {
		final.Status = "cancelled"
		final.DurationMs = time.Since(startedAt).Milliseconds()
		if err := result.Write(filepath.Join(repoRoot, ".watt", "result.json"), final); err != nil {
			return Outcome{Code: ExitInternalError, Result: final}, err
		}
		return Outcome{Code: ExitCancelled, Result: final}, nil
	}

	code := ExitSuccess
	var failure error
	for _, step := range selected.Steps {
		// Cancellation between steps: do not start the next step (§4.2).
		if ctx.Err() != nil {
			final.Status = "cancelled"
			code = ExitCancelled
			break
		}

		stepStartedAt := time.Now()
		runResult := runner.Run(ctx, runner.Step{
			RepoRoot:        repoRoot,
			Name:            step.Name,
			Exec:            step.Exec,
			Args:            step.Args,
			Run:             step.Run,
			Shell:           step.Shell,
			Cwd:             step.Cwd,
			PipelineEnv:     definition.Env,
			StepEnv:         step.Env,
			Stdout:          stdout,
			Stderr:          stderr,
			ConfirmDeadline: options.ConfirmDeadline,
		})

		final.Steps = append(final.Steps, result.Step{
			Name:            step.Name,
			Status:          string(runResult.Status),
			ExitCode:        runResult.ExitCode,
			StartedAt:       stepStartedAt.Format(time.RFC3339Nano),
			DurationMs:      time.Since(stepStartedAt).Milliseconds(),
			ResolvedCommand: runResult.ResolvedCommand,
			OutputTail: result.OutputTail{
				Stdout: runResult.OutputTail.Stdout,
				Stderr: runResult.OutputTail.Stderr,
			},
		})

		if runResult.InternalErr != nil {
			final.Status = "internal_error"
			code = ExitInternalError
			failure = runResult.InternalErr
			break
		}

		switch runResult.Status {
		case runner.StatusSuccess:
			continue
		case runner.StatusEnvironmentUnavailable:
			final.Status = "environment_unavailable"
			code = ExitEnvironmentUnavailable
			failure = runResult.Err
		case runner.StatusFailed:
			final.Status = "failed"
			code = ExitStepFailed
			failure = runResult.Err
		case runner.StatusCancelled:
			final.Status = "cancelled"
			code = ExitCancelled
		default:
			final.Status = "internal_error"
			code = ExitInternalError
			failure = fmt.Errorf("step %q returned unknown status %q", step.Name, runResult.Status)
		}
		break
	}
	final.DurationMs = time.Since(startedAt).Milliseconds()

	resultPath := filepath.Join(repoRoot, ".watt", "result.json")
	if err := result.Write(resultPath, final); err != nil {
		return Outcome{Code: ExitInternalError, Result: final}, err
	}
	return Outcome{Code: code, Result: final}, failure
}

func missingPipelineError(name string, pipelines map[string]pipeline.Pipeline) error {
	names := make([]string, 0, len(pipelines))
	for pipelineName := range pipelines {
		names = append(names, pipelineName)
	}
	sort.Strings(names)
	return fmt.Errorf("pipeline %q not found; available pipelines: %s", name, strings.Join(names, ", "))
}
