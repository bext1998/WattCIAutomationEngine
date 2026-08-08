package main

import (
	"bytes"
	"errors"
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

func TestRootCommandRunAndCheckReportNotImplemented(t *testing.T) {
	for _, name := range []string{"run", "check"} {
		t.Run(name, func(t *testing.T) {
			command := newRootCommand()
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			command.SetOut(&stdout)
			command.SetErr(&stderr)
			command.SetArgs([]string{name})

			if got, want := execute(command), EXIT_INTERNAL_ERROR; got != want {
				t.Errorf("exit code = %d, want %d", got, want)
			}
			if got := stdout.String(); got != "" {
				t.Errorf("stdout = %q, want empty", got)
			}
			if got, want := stderr.String(), name+" is not implemented\n"; got != want {
				t.Errorf("stderr = %q, want %q", got, want)
			}
		})
	}
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
