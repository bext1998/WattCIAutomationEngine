package runner

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestRunExecStepStartsTargetDirectly(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	got := Run(Step{
		Name:    "probe",
		Exec:    os.Args[0],
		Args:    []string{"-test.run=TestRunnerHelperProcess", "--", "direct"},
		HostEnv: environmentFromEntries(append(os.Environ(), "WATT_RUNNER_HELPER=1")),
		Stdout:  &stdout,
		Stderr:  &stderr,
	})

	if got.Status != StatusSuccess {
		t.Fatalf("status = %q, want %q; error = %v", got.Status, StatusSuccess, got.Err)
	}
	if got.ExitCode == nil || *got.ExitCode != 0 {
		t.Fatalf("exit code = %v, want 0", got.ExitCode)
	}
	if got.OutputTail.Stdout != "stdout" {
		t.Errorf("stdout tail = %q, want %q", got.OutputTail.Stdout, "stdout")
	}
	if got.OutputTail.Stderr != "stderr" {
		t.Errorf("stderr tail = %q, want %q", got.OutputTail.Stderr, "stderr")
	}
	if stdout.String() != "stdout" {
		t.Errorf("stdout passthrough = %q, want %q", stdout.String(), "stdout")
	}
	if stderr.String() != "stderr" {
		t.Errorf("stderr passthrough = %q, want %q", stderr.String(), "stderr")
	}
}

func TestRunReportsEnvironmentUnavailableWithoutExitCode(t *testing.T) {
	got := Run(Step{Exec: "watt-runner-command-that-does-not-exist"})

	if got.Status != StatusEnvironmentUnavailable {
		t.Fatalf("status = %q, want %q; error = %v", got.Status, StatusEnvironmentUnavailable, got.Err)
	}
	if got.ExitCode != nil {
		t.Errorf("exit code = %v, want nil", got.ExitCode)
	}
	if !strings.Contains(got.Err.Error(), "watt-runner-command-that-does-not-exist") {
		t.Errorf("error = %q, want missing command name", got.Err)
	}
	if got.OutputTail != (OutputTail{}) {
		t.Errorf("output tail = %#v, want empty output", got.OutputTail)
	}
}

func TestShellStep_CmdSupport(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	tempDir := t.TempDir()
	t.Setenv("TEMP", tempDir)
	t.Setenv("TMP", tempDir)

	got := Run(Step{
		Name:   "cmd",
		Shell:  "cmd",
		Run:    "echo cmd stdout\necho cmd stderr 1>&2\n",
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if got.Status != StatusSuccess {
		t.Fatalf("status = %q, want %q; error = %v", got.Status, StatusSuccess, got.Err)
	}
	if got.ExitCode == nil || *got.ExitCode != 0 {
		t.Fatalf("exit code = %v, want 0", got.ExitCode)
	}
	if !strings.Contains(got.OutputTail.Stdout, "cmd stdout") {
		t.Errorf("stdout tail = %q, want cmd output", got.OutputTail.Stdout)
	}
	if !strings.Contains(got.OutputTail.Stderr, "cmd stderr") {
		t.Errorf("stderr tail = %q, want cmd output", got.OutputTail.Stderr)
	}
	if !strings.Contains(stdout.String(), "cmd stdout") {
		t.Errorf("stdout passthrough = %q, want cmd output", stdout.String())
	}
	if !strings.Contains(stderr.String(), "cmd stderr") {
		t.Errorf("stderr passthrough = %q, want cmd output", stderr.String())
	}
	if matches, err := filepath.Glob(filepath.Join(tempDir, "watt-shell-*.cmd")); err != nil || len(matches) != 0 {
		t.Errorf("temporary cmd scripts = %v, error = %v; want none", matches, err)
	}
}

func TestShellStep_PwshSupport(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	got := Run(Step{
		Name:   "pwsh",
		Shell:  "pwsh",
		Run:    "Write-Output 'pwsh stdout'; [Console]::Error.Write('pwsh stderr')",
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if got.Status != StatusSuccess {
		t.Fatalf("status = %q, want %q; error = %v", got.Status, StatusSuccess, got.Err)
	}
	if got.ExitCode == nil || *got.ExitCode != 0 {
		t.Fatalf("exit code = %v, want 0", got.ExitCode)
	}
	if !strings.Contains(got.OutputTail.Stdout, "pwsh stdout") {
		t.Errorf("stdout tail = %q, want pwsh output", got.OutputTail.Stdout)
	}
	if !strings.Contains(got.OutputTail.Stderr, "pwsh stderr") {
		t.Errorf("stderr tail = %q, want pwsh output", got.OutputTail.Stderr)
	}
	if !strings.Contains(stdout.String(), "pwsh stdout") {
		t.Errorf("stdout passthrough = %q, want pwsh output", stdout.String())
	}
	if !strings.Contains(stderr.String(), "pwsh stderr") {
		t.Errorf("stderr passthrough = %q, want pwsh output", stderr.String())
	}
}

func TestShellStep_PwshFailureReportsExitCode(t *testing.T) {
	got := Run(Step{
		Name:  "pwsh-failure",
		Shell: "pwsh",
		Run:   "exit 3",
	})

	if got.Status != StatusFailed {
		t.Fatalf("status = %q, want %q; error = %v", got.Status, StatusFailed, got.Err)
	}
	if got.ExitCode == nil || *got.ExitCode != 3 {
		t.Errorf("exit code = %v, want 3", got.ExitCode)
	}
}

func TestShellStep_CmdMultilineFailureReportsExitCode(t *testing.T) {
	var stdout bytes.Buffer
	tempDir := t.TempDir()
	t.Setenv("TEMP", tempDir)
	t.Setenv("TMP", tempDir)

	got := Run(Step{
		Name:   "cmd-failure",
		Shell:  "cmd",
		Run:    "echo step-one\necho step-two\nexit 7\n",
		Stdout: &stdout,
	})

	if got.Status != StatusFailed {
		t.Fatalf("status = %q, want %q; error = %v", got.Status, StatusFailed, got.Err)
	}
	if got.ExitCode == nil || *got.ExitCode != 7 {
		t.Errorf("exit code = %v, want 7", got.ExitCode)
	}
	for _, line := range []string{"step-one", "step-two"} {
		if !strings.Contains(got.OutputTail.Stdout, line) {
			t.Errorf("stdout tail = %q, want %q", got.OutputTail.Stdout, line)
		}
		if !strings.Contains(stdout.String(), line) {
			t.Errorf("stdout passthrough = %q, want %q", stdout.String(), line)
		}
	}
	if matches, err := filepath.Glob(filepath.Join(tempDir, "watt-shell-*.cmd")); err != nil || len(matches) != 0 {
		t.Errorf("temporary cmd scripts = %v, error = %v; want none", matches, err)
	}
}

func TestMissingPwsh_NoFallbackTo51(t *testing.T) {
	binDir := t.TempDir()
	marker := filepath.Join(t.TempDir(), "powershell-51-ran")
	fakePowerShell := filepath.Join(binDir, "powershell.cmd")
	if err := os.WriteFile(fakePowerShell, []byte("@echo off\r\necho fallback > \"%WATT_RUNNER_FALLBACK_MARKER%\"\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)
	t.Setenv("PATHEXT", ".COM;.EXE;.BAT;.CMD")
	if _, err := exec.LookPath("powershell"); err != nil {
		t.Fatalf("fallback decoy cannot be resolved: %v", err)
	}

	host := environmentFromEntries(os.Environ())
	host["WATT_RUNNER_FALLBACK_MARKER"] = marker
	got := Run(Step{
		Name:    "missing-pwsh",
		Shell:   "pwsh",
		Run:     "Write-Output should-not-run",
		HostEnv: host,
	})

	if got.Status != StatusEnvironmentUnavailable {
		t.Fatalf("status = %q, want %q; error = %v", got.Status, StatusEnvironmentUnavailable, got.Err)
	}
	if got.ExitCode != nil {
		t.Errorf("exit code = %v, want nil", got.ExitCode)
	}
	if got.Err == nil || !strings.Contains(got.Err.Error(), "pwsh") {
		t.Errorf("error = %v, want missing pwsh", got.Err)
	}
	if _, err := os.Stat(marker); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Windows PowerShell fallback marker exists or could not be checked: %v", err)
	}
}

func TestRunReportsCwdFailureWithoutExitCode(t *testing.T) {
	got := Run(Step{
		Exec: os.Args[0],
		Cwd:  filepath.Join(t.TempDir(), "does-not-exist"),
	})

	if got.Status != StatusFailed {
		t.Fatalf("status = %q, want %q; error = %v", got.Status, StatusFailed, got.Err)
	}
	if got.ExitCode != nil {
		t.Errorf("exit code = %v, want nil", got.ExitCode)
	}
	if got.ResolvedCommand == "" {
		t.Error("resolved command is empty, want resolved command for a cwd failure")
	}
	if got.OutputTail != (OutputTail{}) {
		t.Errorf("output tail = %#v, want empty output", got.OutputTail)
	}
}

func TestRunKeepsOutputTailWithinUtf8ByteLimit(t *testing.T) {
	got := Run(Step{
		Exec:    os.Args[0],
		Args:    []string{"-test.run=TestRunnerHelperProcess"},
		HostEnv: environmentFromEntries(append(os.Environ(), "WATT_RUNNER_HELPER=1", "WATT_RUNNER_LARGE=1")),
	})

	if got.Status != StatusSuccess {
		t.Fatalf("status = %q, want %q; error = %v", got.Status, StatusSuccess, got.Err)
	}
	if len([]byte(got.OutputTail.Stdout)) > maxOutputTailBytes {
		t.Fatalf("stdout tail bytes = %d, want <= %d", len([]byte(got.OutputTail.Stdout)), maxOutputTailBytes)
	}
	if !utf8.ValidString(got.OutputTail.Stdout) {
		t.Fatal("stdout tail is not valid UTF-8")
	}
	if !strings.HasSuffix(got.OutputTail.Stdout, "尾端") {
		t.Errorf("stdout tail = %q, want suffix %q", got.OutputTail.Stdout, "尾端")
	}
}

func TestRunnerHelperProcess(t *testing.T) {
	if os.Getenv("WATT_RUNNER_HELPER") != "1" {
		return
	}
	if os.Getenv("WATT_RUNNER_LARGE") == "1" {
		fmt.Fprint(os.Stdout, strings.Repeat("x", 9000)+"尾端")
		os.Exit(0)
	}
	fmt.Fprint(os.Stdout, "stdout")
	fmt.Fprint(os.Stderr, "stderr")
	os.Exit(0)
}

func environmentFromEntries(entries []string) map[string]string {
	values := make(map[string]string)
	for _, entry := range entries {
		key, value, found := strings.Cut(entry, "=")
		if found {
			values[key] = value
		}
	}
	return values
}
