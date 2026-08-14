package orchestrator

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
