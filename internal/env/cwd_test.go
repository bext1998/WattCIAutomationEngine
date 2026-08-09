package env

import (
	"path/filepath"
	"testing"
)

func TestStep_CwdResolvedRelativeToRepoRoot(t *testing.T) {
	repoRoot := filepath.Join("repo", "root")
	absoluteStepCwd, err := filepath.Abs(filepath.Join("absolute", "step"))
	if err != nil {
		t.Fatalf("filepath.Abs() error = %v", err)
	}

	tests := []struct {
		name    string
		stepCwd string
		want    string
	}{
		{
			name:    "relative cwd is resolved from repo root",
			stepCwd: filepath.Join("build", "work"),
			want:    filepath.Join(repoRoot, "build", "work"),
		},
		{
			name:    "empty cwd uses repo root",
			stepCwd: "",
			want:    repoRoot,
		},
		{
			name:    "absolute cwd is preserved",
			stepCwd: absoluteStepCwd,
			want:    absoluteStepCwd,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := ResolveCwd(repoRoot, test.stepCwd); got != test.want {
				t.Errorf("ResolveCwd(%q, %q) = %q, want %q", repoRoot, test.stepCwd, got, test.want)
			}
		})
	}
}
