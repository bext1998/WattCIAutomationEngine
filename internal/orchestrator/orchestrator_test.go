package orchestrator

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
