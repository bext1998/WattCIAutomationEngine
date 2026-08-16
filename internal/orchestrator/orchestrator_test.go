package orchestrator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bext1998/WattCIAutomationEngine/internal/result"
)

func TestRunPwshShellStepWritesResult(t *testing.T) {
	repoRoot := t.TempDir()
	pipelinePath := filepath.Join(repoRoot, "watt.yaml")
	pipeline := "version: 1\npipelines:\n  default:\n    steps:\n      - name: Package\n        shell: pwsh\n        run: |\n          Write-Output package-complete\n"
	if err := os.WriteFile(pipelinePath, []byte(pipeline), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	outcome, err := Run(Options{
		RepoRoot:     repoRoot,
		PipelinePath: pipelinePath,
		Stdout:       &stdout,
		Stderr:       io.Discard,
	})

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if outcome.Code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d", outcome.Code, ExitSuccess)
	}
	if len(outcome.Result.Steps) != 1 || outcome.Result.Steps[0].Status != "success" {
		t.Fatalf("steps = %#v, want one successful Package step", outcome.Result.Steps)
	}
	if !strings.Contains(stdout.String(), "package-complete") {
		t.Errorf("stdout = %q, want package output", stdout.String())
	}

	contents, err := os.ReadFile(filepath.Join(repoRoot, ".watt", "result.json"))
	if err != nil {
		t.Fatal(err)
	}
	var written result.Result
	if err := json.Unmarshal(contents, &written); err != nil {
		t.Fatal(err)
	}
	if written.Status != "success" || len(written.Steps) != 1 || written.Steps[0].Name != "Package" || written.Steps[0].Status != "success" {
		t.Errorf("result.json = %#v, want successful Package step", written)
	}
}

func TestRunMissingPwshReturnsEnvironmentUnavailable(t *testing.T) {
	repoRoot := t.TempDir()
	pipelinePath := filepath.Join(repoRoot, "watt.yaml")
	pipeline := "version: 1\npipelines:\n  default:\n    steps:\n      - name: script\n        run: Write-Output should-not-run\n"
	if err := os.WriteFile(pipelinePath, []byte(pipeline), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", t.TempDir())

	outcome, err := Run(Options{
		RepoRoot:     repoRoot,
		PipelinePath: pipelinePath,
		Stdout:       io.Discard,
		Stderr:       io.Discard,
	})

	if outcome.Code != ExitEnvironmentUnavailable {
		t.Fatalf("exit code = %d, want %d; error = %v", outcome.Code, ExitEnvironmentUnavailable, err)
	}
	if err == nil || !strings.Contains(err.Error(), "pwsh") {
		t.Errorf("error = %v, want missing pwsh", err)
	}
	if len(outcome.Result.Steps) != 1 || outcome.Result.Steps[0].Status != "environment_unavailable" {
		t.Errorf("steps = %#v, want one environment_unavailable step", outcome.Result.Steps)
	}
}

func TestEnvDiagnostics_NoValuesLeaked(t *testing.T) {
	const canary = "orchestrator-secret-canary"
	repoRoot := t.TempDir()
	pipelinePath := filepath.Join(repoRoot, "watt.yaml")
	pipeline := fmt.Sprintf("version: 1\npipelines:\n  default:\n    steps:\n      - name: canary\n        exec: '%s'\n        args:\n          - -test.run=TestOrchestratorHelperProcess\n        env:\n          WATT_ORCH_HELPER: '1'\n", yamlQuote(os.Args[0]))
	if err := os.WriteFile(pipelinePath, []byte(pipeline), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("WATT_ORCH_CANARY", canary)

	var stdout bytes.Buffer
	outcome, err := Run(Options{
		RepoRoot:     repoRoot,
		PipelinePath: pipelinePath,
		Stdout:       &stdout,
		Stderr:       io.Discard,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if outcome.Code != ExitSuccess {
		t.Fatalf("exit code = %d, want %d", outcome.Code, ExitSuccess)
	}
	if !strings.Contains(stdout.String(), canary) {
		t.Errorf("stdout = %q, want unredacted canary", stdout.String())
	}

	contents, err := os.ReadFile(filepath.Join(repoRoot, ".watt", "result.json"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(contents, []byte(canary)) {
		t.Errorf("result.json contains canary: %s", contents)
	}

	var written result.Result
	if err := json.Unmarshal(contents, &written); err != nil {
		t.Fatal(err)
	}
	if len(written.Environment.ShellAvailable) != 3 {
		t.Errorf("shell_available = %#v, want pwsh/cmd/bash diagnostics", written.Environment.ShellAvailable)
	}
	if !contains(written.Environment.EnvVarNames, "WATT_ORCH_CANARY") {
		t.Errorf("env_var_names = %#v, want WATT_ORCH_CANARY", written.Environment.EnvVarNames)
	}
}

func TestRunCancellationReturnsCancelled(t *testing.T) {
	repoRoot := t.TempDir()
	pipelinePath := filepath.Join(repoRoot, "watt.yaml")
	marker := filepath.Join(repoRoot, "started.marker")

	pipeline := fmt.Sprintf("version: 1\npipelines:\n  default:\n    steps:\n      - name: slow\n        exec: '%s'\n        args:\n          - -test.run=TestOrchestratorHelperProcess\n        env:\n          WATT_ORCH_HELPER: '1'\n          WATT_ORCH_MARKER: '%s'\n", yamlQuote(os.Args[0]), yamlQuote(marker))
	if err := os.WriteFile(pipelinePath, []byte(pipeline), 0o644); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan Outcome, 1)
	go func() {
		outcome, _ := Run(Options{
			RepoRoot:     repoRoot,
			PipelinePath: pipelinePath,
			Stdout:       io.Discard,
			Stderr:       io.Discard,
			Context:      ctx,
		})
		done <- outcome
	}()

	waitForFile(t, marker, 10*time.Second)
	cancel()

	outcome := <-done
	if outcome.Code != ExitCancelled {
		t.Fatalf("exit code = %d, want %d", outcome.Code, ExitCancelled)
	}
	if outcome.Result.Status != "cancelled" {
		t.Errorf("result status = %q, want %q", outcome.Result.Status, "cancelled")
	}
	if len(outcome.Result.Steps) != 1 || outcome.Result.Steps[0].Status != "cancelled" {
		t.Errorf("steps = %#v, want one cancelled step", outcome.Result.Steps)
	}
}

func TestRunCancellationUnconfirmedReturnsInternalError(t *testing.T) {
	repoRoot := t.TempDir()
	pipelinePath := filepath.Join(repoRoot, "watt.yaml")
	marker := filepath.Join(repoRoot, "started.marker")

	pipeline := fmt.Sprintf("version: 1\npipelines:\n  default:\n    steps:\n      - name: slow\n        exec: '%s'\n        args:\n          - -test.run=TestOrchestratorHelperProcess\n        env:\n          WATT_ORCH_HELPER: '1'\n          WATT_ORCH_MARKER: '%s'\n", yamlQuote(os.Args[0]), yamlQuote(marker))
	if err := os.WriteFile(pipelinePath, []byte(pipeline), 0o644); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan Outcome, 1)
	go func() {
		outcome, _ := Run(Options{
			RepoRoot:     repoRoot,
			PipelinePath: pipelinePath,
			Stdout:       io.Discard,
			Stderr:       io.Discard,
			Context:      ctx,
			// Too short to confirm the termination, forcing internal_error.
			ConfirmDeadline: time.Nanosecond,
		})
		done <- outcome
	}()

	waitForFile(t, marker, 10*time.Second)
	cancel()

	outcome := <-done
	if outcome.Code != ExitInternalError {
		t.Fatalf("exit code = %d, want %d", outcome.Code, ExitInternalError)
	}
	if outcome.Result.Status != "internal_error" {
		t.Errorf("result status = %q, want %q", outcome.Result.Status, "internal_error")
	}
	if len(outcome.Result.Steps) != 1 || outcome.Result.Steps[0].Status != "cancelled" {
		t.Errorf("steps = %#v, want one cancelled step", outcome.Result.Steps)
	}
}

func yamlQuote(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}

func waitForFile(t *testing.T, path string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("file %q not created within %s", path, timeout)
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func TestOrchestratorHelperProcess(t *testing.T) {
	if os.Getenv("WATT_ORCH_HELPER") != "1" {
		return
	}
	if canary := os.Getenv("WATT_ORCH_CANARY"); canary != "" {
		fmt.Fprint(os.Stdout, canary)
		os.Exit(0)
	}
	if marker := os.Getenv("WATT_ORCH_MARKER"); marker != "" {
		_ = os.WriteFile(marker, []byte("started"), 0o644)
		time.Sleep(60 * time.Second)
		os.Exit(0)
	}
	os.Exit(0)
}
