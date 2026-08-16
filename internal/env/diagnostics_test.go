package env

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestProbeDiagnostics_ReportsSafeHostEnvironment(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	t.Setenv("WATT_ENV_DIAGNOSTIC_CANARY", "environment-secret-canary")

	diagnostics := ProbeDiagnostics("windows", "amd64", []string{os.Args[0], "missing-tool"})

	if diagnostics.OS != "windows" || diagnostics.Arch != "amd64" {
		t.Errorf("OS/Arch = %q/%q, want windows/amd64", diagnostics.OS, diagnostics.Arch)
	}
	for _, shell := range []string{"pwsh", "cmd", "bash"} {
		if available, exists := diagnostics.ShellAvailable[shell]; !exists || available != false {
			t.Errorf("shell_available[%q] = %#v, want false", shell, available)
		}
	}
	if path := diagnostics.ResolvedTools[os.Args[0]]; path == "" || !filepath.IsAbs(path) {
		t.Errorf("resolved_tools[%q] = %q, want absolute executable path", os.Args[0], path)
	}
	if _, exists := diagnostics.ResolvedTools["missing-tool"]; exists {
		t.Error("resolved_tools contains missing tool")
	}
	if !contains(diagnostics.EnvVarNames, "WATT_ENV_DIAGNOSTIC_CANARY") {
		t.Errorf("env_var_names = %#v, want canary name", diagnostics.EnvVarNames)
	}

	contents, err := json.Marshal(diagnostics)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(contents), "environment-secret-canary") {
		t.Errorf("diagnostics contains environment value: %s", contents)
	}
}

func TestProbeDiagnostics_PwshVersionTimeoutFallsBack(t *testing.T) {
	binDir := t.TempDir()
	fakePwsh := filepath.Join(binDir, "pwsh.cmd")
	if err := os.WriteFile(fakePwsh, []byte("@echo off\r\n:loop\r\ngoto loop\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)
	t.Setenv("PATHEXT", ".COM;.EXE;.BAT;.CMD")

	previousTimeout := pwshVersionProbeTimeout
	pwshVersionProbeTimeout = 20 * time.Millisecond
	t.Cleanup(func() { pwshVersionProbeTimeout = previousTimeout })

	started := time.Now()
	diagnostics := ProbeDiagnostics("windows", "amd64", nil)
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("ProbeDiagnostics() took %s, want pwsh version probe to time out promptly", elapsed)
	}
	if available := diagnostics.ShellAvailable["pwsh"]; available != true {
		t.Errorf("shell_available[pwsh] = %#v, want true fallback after timeout", available)
	}
}

func TestProbeDiagnostics_PwshVersionTimeoutWithGrandchildFallsBack(t *testing.T) {
	binDir := t.TempDir()
	marker := filepath.Join(t.TempDir(), "grandchild-started")
	fakePwsh := filepath.Join(binDir, "pwsh.cmd")
	contents := "@echo off\r\nstart \"\" /b cmd /c \"ping -n 4 127.0.0.1 > nul\"\r\necho started > \"%WATT_PWSH_GRANDCHILD_MARKER%\"\r\n:loop\r\ngoto loop\r\n"
	if err := os.WriteFile(fakePwsh, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
	systemRoot := os.Getenv("SystemRoot")
	if systemRoot == "" {
		t.Fatal("SystemRoot is empty")
	}
	t.Setenv("PATH", binDir+";"+filepath.Join(systemRoot, "System32"))
	t.Setenv("PATHEXT", ".COM;.EXE;.BAT;.CMD")
	t.Setenv("WATT_PWSH_GRANDCHILD_MARKER", marker)

	previousTimeout := pwshVersionProbeTimeout
	pwshVersionProbeTimeout = 100 * time.Millisecond
	t.Cleanup(func() { pwshVersionProbeTimeout = previousTimeout })

	started := time.Now()
	diagnostics := ProbeDiagnostics("windows", "amd64", nil)
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("ProbeDiagnostics() took %s, want timeout despite a grandchild holding stdout", elapsed)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("grandchild marker %q was not written: %v", marker, err)
	}
	if available := diagnostics.ShellAvailable["pwsh"]; available != true {
		t.Errorf("shell_available[pwsh] = %#v, want true fallback after timeout", available)
	}
}

func TestRedactKnownValues(t *testing.T) {
	got := RedactKnownValues("long-secret-value short 1234567", map[string]string{
		"LONG":  "long-secret-value",
		"SHORT": "1234567",
	})
	if got != "[REDACTED] short 1234567" {
		t.Errorf("RedactKnownValues() = %q", got)
	}
}

func TestRedactKnownValues_RedactsOverlappingValues(t *testing.T) {
	const (
		first  = "abcdefghXY"
		second = "XYijklmnop"
	)
	got := RedactKnownValues(first+"ijklmnop", map[string]string{
		"TOKEN_A": first,
		"TOKEN_B": second,
	})
	if got != redactedValue {
		t.Errorf("RedactKnownValues() = %q, want a single redacted span", got)
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
