package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCommandVersionOutput(t *testing.T) {
	command := newRootCommand()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.SetOut(&stdout)
	command.SetErr(&stderr)
	command.SetArgs([]string{"--version"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if got, want := stdout.String(), "watt dev\n"; got != want {
		t.Errorf("version output = %q, want %q", got, want)
	}
	if got := stderr.String(); got != "" {
		t.Errorf("version stderr = %q, want empty", got)
	}
}

func TestRootCommandRunReportsNotImplemented(t *testing.T) {
	command := newRootCommand()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.SetOut(&stdout)
	command.SetErr(&stderr)
	command.SetArgs([]string{"run"})

	if got, want := execute(command), EXIT_INTERNAL_ERROR; got != want {
		t.Errorf("exit code = %d, want %d", got, want)
	}
	if got := stdout.String(); got != "" {
		t.Errorf("stdout = %q, want empty", got)
	}
	if got, want := stderr.String(), "run is not implemented\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

func TestCheck_NoSideEffects(t *testing.T) {
	workdir := t.TempDir()
	oldWorkingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(workdir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldWorkingDirectory); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	resultDirectory := filepath.Join(workdir, ".watt")
	if err := os.Mkdir(resultDirectory, 0o755); err != nil {
		t.Fatal(err)
	}
	resultPath := filepath.Join(resultDirectory, "result.json")
	resultContents := []byte(`{"status":"existing"}`)
	if err := os.WriteFile(resultPath, resultContents, 0o644); err != nil {
		t.Fatal(err)
	}
	pipelineContents := []byte("version: 1\npipelines:\n  default:\n    steps:\n      - name: destructive\n        exec: cmd.exe\n        args:\n          - /c\n          - echo started>started.marker\n")
	if err := os.WriteFile(filepath.Join(workdir, "watt.yaml"), pipelineContents, 0o644); err != nil {
		t.Fatal(err)
	}

	before := snapshotFiles(t, workdir)
	command := newRootCommand()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.SetOut(&stdout)
	command.SetErr(&stderr)
	command.SetArgs([]string{"check"})

	if got, want := execute(command), EXIT_SUCCESS; got != want {
		t.Fatalf("exit code = %d, want %d; stderr = %q", got, want, stderr.String())
	}
	if got := stdout.String(); got != "" {
		t.Errorf("stdout = %q, want empty", got)
	}
	if got := stderr.String(); got != "" {
		t.Errorf("stderr = %q, want empty", got)
	}

	after := snapshotFiles(t, workdir)
	if !reflect.DeepEqual(after, before) {
		t.Errorf("check changed the filesystem: before=%v after=%v", before, after)
	}
}

func TestCheck_FailurePaths(t *testing.T) {
	tests := []struct {
		name     string
		pipeline string
	}{
		{
			name: "load failure",
		},
		{
			name:     "validation failure",
			pipeline: "version: 1\npipelines:\n  default:\n    steps:\n      - name: invalid\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			workdir := t.TempDir()
			oldWorkingDirectory, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Chdir(workdir); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := os.Chdir(oldWorkingDirectory); err != nil {
					t.Errorf("restore working directory: %v", err)
				}
			})

			if test.pipeline != "" {
				if err := os.WriteFile(filepath.Join(workdir, "watt.yaml"), []byte(test.pipeline), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			command := newRootCommand()
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			command.SetOut(&stdout)
			command.SetErr(&stderr)
			command.SetArgs([]string{"check"})

			if got, want := execute(command), EXIT_INVALID_PIPELINE; got != want {
				t.Errorf("exit code = %d, want %d", got, want)
			}
			if got := stdout.String(); got != "" {
				t.Errorf("stdout = %q, want empty", got)
			}
			if got := stderr.String(); got == "" {
				t.Error("stderr is empty, want an error message")
			}
		})
	}
}

type snapshotEntry struct {
	directory bool
	contents  []byte
}

func snapshotFiles(t *testing.T, root string) map[string]snapshotEntry {
	t.Helper()
	snapshot := make(map[string]snapshotEntry)
	if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relativePath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			snapshot[relativePath] = snapshotEntry{directory: true}
			return nil
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		snapshot[relativePath] = snapshotEntry{contents: contents}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return snapshot
}

func TestRootCommandUsageErrorsShareNonRetryableInputExitCode(t *testing.T) {
	if EXIT_USAGE != EXIT_INVALID_PIPELINE {
		t.Fatalf("EXIT_USAGE = %d, want EXIT_INVALID_PIPELINE (%d)", EXIT_USAGE, EXIT_INVALID_PIPELINE)
	}

	tests := []struct {
		name         string
		args         []string
		errorMessage string
	}{
		{name: "unknown command", args: []string{"bogus"}, errorMessage: `unknown command "bogus"`},
		{name: "unknown flag", args: []string{"--bogus"}, errorMessage: "unknown flag: --bogus"},
		{name: "run unexpected arguments", args: []string{"run", "foo", "bar"}, errorMessage: `unknown command "foo"`},
		{name: "check unexpected arguments", args: []string{"check", "foo"}, errorMessage: `unknown command "foo"`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command := newRootCommand()
			var stderr bytes.Buffer
			command.SetErr(&stderr)
			command.SetArgs(test.args)

			if got := execute(command); got != EXIT_USAGE {
				t.Errorf("exit code = %d, want EXIT_USAGE (%d)", got, EXIT_USAGE)
			}
			if got := strings.Count(stderr.String(), test.errorMessage); got != 1 {
				t.Errorf("stderr contains %q %d times, want exactly once: %q", test.errorMessage, got, stderr.String())
			}
		})
	}
}

func TestExitErrorPreservesCodeAndCause(t *testing.T) {
	cause := errors.New("underlying failure")
	err := &exitError{code: EXIT_INTERNAL_ERROR, err: cause}

	if got, want := err.Error(), cause.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
	if !errors.Is(err, cause) {
		t.Error("errors.Is() did not find the wrapped cause")
	}
}

func TestExecuteUnclassifiedErrorUsesInternalErrorClass(t *testing.T) {
	command := &cobra.Command{
		Use: "watt",
		RunE: func(*cobra.Command, []string) error {
			return errors.New("unexpected failure")
		},
	}
	var stderr bytes.Buffer
	command.SetErr(&stderr)

	if got, want := execute(command), EXIT_INTERNAL_ERROR; got != want {
		t.Errorf("unclassified error exit code = %d, want EXIT_INTERNAL_ERROR (%d)", got, want)
	}
	if !strings.Contains(stderr.String(), "unexpected failure") {
		t.Errorf("stderr = %q, want the unclassified error", stderr.String())
	}
}

func TestRootCommandHelpExcludesCompletion(t *testing.T) {
	command := newRootCommand()
	var stdout bytes.Buffer
	command.SetOut(&stdout)
	command.SetArgs([]string{"--help"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	help := stdout.String()
	for _, name := range []string{"check", "run"} {
		if !strings.Contains(help, "  "+name) {
			t.Errorf("help output does not list %q: %q", name, help)
		}
	}
	if strings.Contains(help, "  completion") {
		t.Errorf("help output exposes completion command: %q", help)
	}
}
