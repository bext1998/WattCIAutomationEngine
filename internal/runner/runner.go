package runner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/bext1998/WattCIAutomationEngine/internal/env"
	"github.com/bext1998/WattCIAutomationEngine/internal/proc"
)

const maxOutputTailBytes = 8192

// Status is the step-level status reported by a single step execution.
type Status string

const (
	StatusSuccess                Status = "success"
	StatusFailed                 Status = "failed"
	StatusEnvironmentUnavailable Status = "environment_unavailable"
	StatusCancelled              Status = "cancelled"
)

// Step is the execution-time input for one pipeline step.
type Step struct {
	RepoRoot    string
	Name        string
	Exec        string
	Args        []string
	Run         string
	Shell       string
	Cwd         string
	HostEnv     map[string]string
	PipelineEnv map[string]string
	StepEnv     map[string]string
	Stdout      io.Writer
	Stderr      io.Writer
}

type OutputTail struct {
	Stdout string
	Stderr string
}

type Result struct {
	Name            string
	Status          Status
	ExitCode        *int
	ResolvedCommand string
	OutputTail      OutputTail
	Err             error

	// InternalErr is non-nil when the step itself finished in a way that must
	// escalate the pipeline's top-level status to "internal_error" (for
	// example, a cancelled process tree that could not be confirmed terminated
	// within the deadline). The step-level Status still follows the step status
	// enum (cancelled / failed); InternalErr is what the orchestrator consults.
	InternalErr error
}

func Run(ctx context.Context, step Step) Result {
	result := Result{
		Name: step.Name,
		OutputTail: OutputTail{
			Stdout: "",
			Stderr: "",
		},
	}

	effectiveEnvironment := step.HostEnv
	if effectiveEnvironment == nil {
		effectiveEnvironment = hostEnvironment()
	}
	effectiveEnvironment = env.Merge(effectiveEnvironment, step.PipelineEnv, step.StepEnv)

	commandName := step.Exec
	commandArgs := step.Args
	commandPath := step.Exec
	if strings.TrimSpace(step.Run) != "" {
		resolvedShell, err := env.ResolveExecutable(step.Shell)
		if err != nil {
			result.Status = StatusEnvironmentUnavailable
			result.Err = fmt.Errorf("shell %q is unavailable: %w", step.Shell, err)
			return result
		}
		commandName = step.Shell
		if step.Shell == "cmd" {
			var cleanup func()
			commandArgs, cleanup, err = cmdArgs(step.Run)
			if err != nil {
				result.Status = StatusFailed
				result.Err = fmt.Errorf("create cmd script: %w", err)
				return result
			}
			defer cleanup()
		} else {
			commandArgs = shellArgs(step.Shell, step.Run)
		}
		commandPath = resolvedShell
	} else {
		resolved, err := exec.LookPath(step.Exec)
		if err != nil {
			result.Status = StatusEnvironmentUnavailable
			result.Err = fmt.Errorf("command %q is unavailable: %w", commandName, err)
			return result
		}
		commandPath = resolved
	}

	result.ResolvedCommand = strings.Join(append([]string{commandName}, commandArgs...), " ")

	stdoutTail := &tailBuffer{}
	stderrTail := &tailBuffer{}
	stdout := step.Stdout
	if stdout == nil {
		stdout = io.Discard
	}
	stderr := step.Stderr
	if stderr == nil {
		stderr = io.Discard
	}

	outcome := proc.Run(ctx, proc.Spec{
		Path:   commandPath,
		Args:   commandArgs,
		Env:    environmentEntries(effectiveEnvironment),
		Dir:    env.ResolveCwd(step.RepoRoot, step.Cwd),
		Stdout: io.MultiWriter(stdout, stdoutTail),
		Stderr: io.MultiWriter(stderr, stderrTail),
	})

	result.OutputTail = OutputTail{Stdout: stdoutTail.String(), Stderr: stderrTail.String()}

	switch outcome.Status {
	case proc.StatusStartFailed:
		if outcome.NotFound {
			result.Status = StatusEnvironmentUnavailable
			result.Err = fmt.Errorf("command %q is unavailable: %w", commandName, outcome.Err)
		} else {
			result.Status = StatusFailed
			result.Err = fmt.Errorf("start command %q: %w", commandName, outcome.Err)
		}
	case proc.StatusExited:
		exitCode := outcome.ExitCode
		result.ExitCode = &exitCode
		if exitCode == 0 {
			result.Status = StatusSuccess
		} else {
			result.Status = StatusFailed
			result.Err = fmt.Errorf("command %q failed with exit code %d", commandName, exitCode)
		}
	case proc.StatusCancelled:
		result.Status = StatusCancelled
	case proc.StatusInternalError:
		result.InternalErr = outcome.Err
		if ctx.Err() != nil {
			// The step was cancelled but its tree could not be confirmed
			// terminated; the step stays "cancelled" (R-7) and the pipeline
			// escalates to internal_error via InternalErr.
			result.Status = StatusCancelled
		} else {
			// A non-cancellation internal error (e.g. orphan cleanup on a
			// normally-completed step could not be confirmed).
			result.Status = StatusFailed
		}
	}

	return result
}

func shellArgs(shell, script string) []string {
	switch shell {
	case "pwsh":
		return []string{"-Command", script}
	default:
		return nil
	}
}

func cmdArgs(script string) ([]string, func(), error) {
	file, err := os.CreateTemp("", "watt-shell-*.cmd")
	if err != nil {
		return nil, nil, err
	}
	path := file.Name()
	cleanup := func() {
		_ = os.Remove(path)
	}

	script = strings.ReplaceAll(script, "\r\n", "\n")
	script = strings.ReplaceAll(script, "\n", "\r\n")
	if _, err := file.WriteString(script); err != nil {
		_ = file.Close()
		cleanup()
		return nil, nil, err
	}
	if err := file.Close(); err != nil {
		cleanup()
		return nil, nil, err
	}

	return []string{"/c", "call", path}, cleanup, nil
}

type tailBuffer struct {
	data []byte
}

func hostEnvironment() map[string]string {
	host := make(map[string]string)
	for _, entry := range os.Environ() {
		key, value, found := strings.Cut(entry, "=")
		if found {
			host[key] = value
		}
	}
	return host
}

func environmentEntries(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	entries := make([]string, 0, len(keys))
	for _, key := range keys {
		entries = append(entries, key+"="+values[key])
	}
	return entries
}

func (buffer *tailBuffer) Write(data []byte) (int, error) {
	buffer.data = append(buffer.data, data...)
	if len(buffer.data) > maxOutputTailBytes+utf8.UTFMax {
		buffer.data = buffer.data[len(buffer.data)-(maxOutputTailBytes+utf8.UTFMax):]
	}
	return len(data), nil
}

func (buffer *tailBuffer) String() string {
	data := buffer.data
	if len(data) > maxOutputTailBytes {
		data = data[len(data)-maxOutputTailBytes:]
	}
	data = bytes.ToValidUTF8(data, []byte("\uFFFD"))
	for len(data) > maxOutputTailBytes {
		data = data[1:]
		data = bytes.ToValidUTF8(data, []byte("\uFFFD"))
	}
	return string(data)
}
