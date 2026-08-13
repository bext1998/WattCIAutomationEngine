package result

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const SchemaVersion = 1

type Result struct {
	SchemaVersion int         `json:"schema_version"`
	Pipeline      string      `json:"pipeline"`
	Status        string      `json:"status"`
	Steps         []Step      `json:"steps"`
	StartedAt     string      `json:"started_at"`
	DurationMs    int64       `json:"duration_ms"`
	WattVersion   string      `json:"watt_version"`
	Environment   Environment `json:"environment"`
}

type Step struct {
	Name            string     `json:"name"`
	Status          string     `json:"status"`
	ExitCode        *int       `json:"exit_code"`
	StartedAt       string     `json:"started_at"`
	DurationMs      int64      `json:"duration_ms"`
	ResolvedCommand string     `json:"resolved_command"`
	OutputTail      OutputTail `json:"output_tail"`
}

type OutputTail struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
}

type Environment struct {
	OS             string            `json:"os"`
	Arch           string            `json:"arch"`
	ShellAvailable map[string]any    `json:"shell_available,omitempty"`
	ResolvedTools  map[string]string `json:"resolved_tools,omitempty"`
	EnvVarNames    []string          `json:"env_var_names,omitempty"`
}

func Write(path string, value Result) error {
	contents, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal result: %w", err)
	}
	contents = append(contents, '\n')

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create result directory: %w", err)
	}
	if err := os.WriteFile(path, contents, 0o644); err != nil {
		return fmt.Errorf("write result %q: %w", path, err)
	}
	return nil
}
