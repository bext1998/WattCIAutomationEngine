package main

import (
	"bytes"
	"strings"
	"testing"
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

func TestRootCommandRunAndCheckAreEmptyStubs(t *testing.T) {
	for _, name := range []string{"run", "check"} {
		t.Run(name, func(t *testing.T) {
			command := newRootCommand()
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			command.SetOut(&stdout)
			command.SetErr(&stderr)
			command.SetArgs([]string{name})

			if err := command.Execute(); err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if got := stdout.String(); got != "" {
				t.Errorf("stdout = %q, want empty", got)
			}
			if got := stderr.String(); got != "" {
				t.Errorf("stderr = %q, want empty", got)
			}
		})
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
