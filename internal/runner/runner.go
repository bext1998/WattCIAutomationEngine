package runner

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/bext1998/WattCIAutomationEngine/internal/env"
)

const maxOutputTailBytes = 8192

type Status string

const (
	StatusSuccess                Status = "success"
	StatusFailed                 Status = "failed"
	StatusEnvironmentUnavailable Status = "environment_unavailable"
)

// Step is the execution-time input for one pipeline step.
type Step struct {
	RepoRoot    string
	Name        string
	Exec        string
	Args        []string
	Run         string
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
}

func Run(step Step) Result {
	result := Result{
		Name: step.Name,
		OutputTail: OutputTail{
			Stdout: "",
			Stderr: "",
		},
	}

	if strings.TrimSpace(step.Run) != "" {
		result.Status = StatusFailed
		result.Err = errors.New("shell step execution is not implemented")
		return result
	}

	effectiveEnvironment := step.HostEnv
	if effectiveEnvironment == nil {
		effectiveEnvironment = hostEnvironment()
	}
	effectiveEnvironment = env.Merge(effectiveEnvironment, step.PipelineEnv, step.StepEnv)

	command := exec.Command(step.Exec, step.Args...)
	command.Env = environmentEntries(effectiveEnvironment)
	command.Dir = env.ResolveCwd(step.RepoRoot, step.Cwd)

	if command.Err != nil {
		result.Status = StatusEnvironmentUnavailable
		result.Err = fmt.Errorf("command %q is unavailable: %w", step.Exec, command.Err)
		return result
	}
	result.ResolvedCommand = strings.Join(append([]string{step.Exec}, step.Args...), " ")

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
	command.Stdout = io.MultiWriter(stdout, stdoutTail)
	command.Stderr = io.MultiWriter(stderr, stderrTail)

	if err := command.Start(); err != nil {
		result.Status = StatusFailed
		if errors.Is(err, exec.ErrNotFound) {
			result.Status = StatusEnvironmentUnavailable
			result.Err = fmt.Errorf("command %q is unavailable: %w", step.Exec, err)
		} else {
			result.Err = fmt.Errorf("start command %q: %w", step.Exec, err)
		}
		result.OutputTail = OutputTail{Stdout: stdoutTail.String(), Stderr: stderrTail.String()}
		return result
	}

	err := command.Wait()
	result.OutputTail = OutputTail{Stdout: stdoutTail.String(), Stderr: stderrTail.String()}
	if command.ProcessState != nil {
		exitCode := command.ProcessState.ExitCode()
		result.ExitCode = &exitCode
	}
	if err != nil {
		result.Status = StatusFailed
		result.Err = fmt.Errorf("command %q failed: %w", step.Exec, err)
		return result
	}

	result.Status = StatusSuccess
	return result
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
