package env

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveExecutable_ResolvesCommand(t *testing.T) {
	systemRoot := os.Getenv("SystemRoot")
	if systemRoot == "" {
		t.Fatal("SystemRoot is empty")
	}
	t.Setenv("PATH", filepath.Join(systemRoot, "System32"))

	path, err := ResolveExecutable("cmd.exe")
	if err != nil {
		t.Fatalf("ResolveExecutable() error = %v, want nil", err)
	}
	if path == "" {
		t.Error("ResolveExecutable() path is empty, want a resolved path")
	}
}

func TestResolveExecutable_UsesPATHEXT(t *testing.T) {
	binDir := t.TempDir()
	commandPath := filepath.Join(binDir, "watt-probe.cmd")
	if err := os.WriteFile(commandPath, []byte("@echo off\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)
	t.Setenv("PATHEXT", ".COM;.EXE;.BAT;.CMD")

	path, err := ResolveExecutable("watt-probe")
	if err != nil {
		t.Fatalf("ResolveExecutable() error = %v", err)
	}
	if !strings.EqualFold(path, commandPath) {
		t.Errorf("ResolveExecutable() = %q, want %q", path, commandPath)
	}
}

func TestResolveExecutable_MissingNameIncludesErrorContext(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	const name = "watt-test-resolve-executable-missing"

	_, err := ResolveExecutable(name)
	if err == nil {
		t.Fatal("ResolveExecutable() error = nil, want an executable lookup error")
	}
	if !strings.Contains(err.Error(), name) {
		t.Errorf("ResolveExecutable() error = %q, want it to contain %q", err, name)
	}
}
