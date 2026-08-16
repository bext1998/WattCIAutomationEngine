package pipeline

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestValidate_ExecAndRunBothSet(t *testing.T) {
	file := validPipelineFile()
	file.Pipelines["default"].Steps[0].Exec = "tool"
	file.Pipelines["default"].Steps[0].Run = "echo hello"

	if err := file.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want an exec/run mutual exclusion error")
	}
}

func TestValidate_NeitherExecNorRun(t *testing.T) {
	file := validPipelineFile()
	file.Pipelines["default"].Steps[0].Exec = ""
	file.Pipelines["default"].Steps[0].Run = "  "

	if err := file.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want an exec/run requirement error")
	}
}

func TestValidate_DuplicateStepName(t *testing.T) {
	file := validPipelineFile()
	definition := file.Pipelines["default"]
	definition.Steps = append(definition.Steps, Step{
		Name: " build",
		Exec: "tool-two",
	})
	file.Pipelines["default"] = definition

	if err := file.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want duplicate step name error")
	}
}

func TestSchemaVersion_IsInteger(t *testing.T) {
	path := writePipeline(t, "version: 1.0\npipelines: {}\n")

	if _, err := Load(path); err == nil {
		t.Fatal("Load() error = nil, want a non-integer schema version error")
	}
}

func TestValidate_StructuralRules(t *testing.T) {
	tests := []struct {
		name string
		file PipelineFile
	}{
		{
			name: "no pipelines",
			file: PipelineFile{Version: 1},
		},
		{
			name: "blank pipeline name",
			file: PipelineFile{Version: 1, Pipelines: map[string]Pipeline{" ": {Steps: []Step{{Name: "step", Exec: "tool"}}}}},
		},
		{
			name: "no steps",
			file: PipelineFile{Version: 1, Pipelines: map[string]Pipeline{"default": {}}},
		},
		{
			name: "blank step name",
			file: PipelineFile{Version: 1, Pipelines: map[string]Pipeline{"default": {Steps: []Step{{Name: " \t", Exec: "tool"}}}}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.file.Validate(); err == nil {
				t.Fatal("Validate() error = nil, want a structural validation error")
			}
		})
	}
}

func TestValidate_FieldModesAndShells(t *testing.T) {
	tests := []struct {
		name      string
		step      Step
		wantError string
	}{
		{
			name:      "whitespace shell with exec",
			step:      Step{Name: "step", Exec: "tool", Shell: "  "},
			wantError: "shell",
		},
		{
			name:      "args with run",
			step:      Step{Name: "step", Run: "echo hello", Args: []string{"unexpected"}},
			wantError: "args",
		},
		{
			name:      "shell with exec",
			step:      Step{Name: "step", Exec: "tool", Shell: "cmd"},
			wantError: "shell",
		},
		{
			name:      "bash is not supported",
			step:      Step{Name: "step", Run: "echo hello", Shell: "bash"},
			wantError: "尚未支援",
		},
		{
			name:      "unknown shell",
			step:      Step{Name: "step", Run: "echo hello", Shell: "zsh"},
			wantError: "unsupported shell",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			file := validPipelineFile()
			file.Pipelines["default"].Steps[0] = test.step
			err := file.Validate()
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("Validate() error = %v, want message containing %q", err, test.wantError)
			}
		})
	}
}

func TestLoad_DefaultsRunShellToPwsh(t *testing.T) {
	path := writePipeline(t, "version: 1\npipelines:\n  default:\n    steps:\n      - name: script\n        run: echo hello\n")

	file, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := file.Pipelines["default"].Steps[0].Shell, "pwsh"; got != want {
		t.Errorf("Shell = %q, want %q", got, want)
	}
	if err := file.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}

func TestLoad_SyntaxError(t *testing.T) {
	path := writePipeline(t, "version: [\npipelines: {}\n")

	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() error = nil, want YAML syntax error")
	}
	if !strings.Contains(err.Error(), "line") {
		t.Errorf("Load() error = %q, want a line number", err)
	}
}

func TestLoad_NullDocument(t *testing.T) {
	for _, contents := range []string{"null\n", "~\n"} {
		t.Run(strings.TrimSpace(contents), func(t *testing.T) {
			path := writePipeline(t, contents)

			_, err := Load(path)
			if err == nil || !strings.Contains(err.Error(), "empty YAML document") {
				t.Fatalf("Load() error = %v, want empty YAML document error", err)
			}
		})
	}
}

func TestLoad_UnknownField(t *testing.T) {
	tests := []struct {
		name     string
		contents string
	}{
		{
			name:     "top level",
			contents: "version: 1\nunknown: true\npipelines: {}\n",
		},
		{
			name:     "pipeline",
			contents: "version: 1\npipelines:\n  default:\n    steps: []\n    unknown: true\n",
		},
		{
			name:     "step",
			contents: "version: 1\npipelines:\n  default:\n    steps:\n      - name: build\n        exec: tool\n        unknown: true\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Load(writePipeline(t, test.contents))
			if err == nil {
				t.Fatal("Load() error = nil, want unknown field error")
			}
			if !strings.Contains(err.Error(), "field") {
				t.Errorf("Load() error = %q, want unknown field context", err)
			}
		})
	}
}

func TestLoad_RejectsNonStringStringFields(t *testing.T) {
	tests := []struct {
		name     string
		contents string
		want     string
	}{
		{
			name:     "steps string",
			contents: "version: 1\npipelines:\n  default:\n    steps: not-a-list\n",
			want:     "cannot unmarshal",
		},
		{
			name:     "pipeline env list",
			contents: "version: 1\nenv: [not-a-map]\npipelines: {}\n",
			want:     "cannot unmarshal",
		},
		{
			name:     "args number",
			contents: "version: 1\npipelines:\n  default:\n    steps:\n      - name: build\n        exec: tool\n        args: [1]\n",
			want:     "args",
		},
		{
			name:     "pipeline env number",
			contents: "version: 1\nenv: {ANSWER: 42}\npipelines: {}\n",
			want:     "env value",
		},
		{
			name:     "step env boolean",
			contents: "version: 1\npipelines:\n  default:\n    steps:\n      - name: build\n        exec: tool\n        env: {CI: true}\n",
			want:     "env value",
		},
		{
			name:     "step name number",
			contents: "version: 1\npipelines:\n  default:\n    steps:\n      - name: 1\n        exec: tool\n",
			want:     "name",
		},
		{
			name:     "pipeline name number",
			contents: "version: 1\npipelines:\n 1:\n    steps: []\n",
			want:     "pipeline name",
		},
		{
			name:     "exec boolean",
			contents: "version: 1\npipelines:\n  default:\n    steps:\n      - name: build\n        exec: true\n",
			want:     "exec",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Load(writePipeline(t, test.contents))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Load() error = %v, want type error containing %q", err, test.want)
			}
		})
	}
}

func TestLoad_DuplicateVersionKeysRejected(t *testing.T) {
	_, err := Load(writePipeline(t, "version: 1\nversion: bad\npipelines: {}\n"))
	if err == nil || !strings.Contains(err.Error(), "already defined") {
		t.Fatalf("Load() error = %v, want duplicate key error", err)
	}
}

func TestLoad_MultiDocumentRejectsEmptyTrailingDocument(t *testing.T) {
	for _, contents := range []string{
		"version: 1\npipelines: {}\n---\n",
		"version: 1\npipelines: {}\n---\n# intentionally empty\n\n",
	} {
		_, err := Load(writePipeline(t, contents))
		if err == nil || !strings.Contains(err.Error(), "multiple YAML documents") {
			t.Fatalf("Load() error = %v, want multiple document error", err)
		}
	}
}

func TestLoad_AnchorsAndMergesRemainStrict(t *testing.T) {
	tests := []struct {
		name     string
		contents string
		want     string
	}{
		{
			name:     "merged unknown field",
			contents: "version: 1\npipelines:\n  default:\n    steps:\n      - &template {unknown: value}\n      - <<: *template\n        name: build\n        exec: tool\n",
			want:     "field unknown",
		},
		{
			name:     "aliased numeric argument",
			contents: "version: &number 1\npipelines:\n  default:\n    steps:\n      - name: build\n        exec: tool\n        args: [*number]\n",
			want:     "args",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Load(writePipeline(t, test.contents))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Load() error = %v, want error containing %q", err, test.want)
			}
		})
	}
}

func TestLoad_CircularMergeKeyRejectedWithoutCrash(t *testing.T) {
	const helperEnv = "WATT_PIPELINE_CIRCULAR_MERGE_HELPER"
	contents := "version: 1\npipelines:\n  default:\n    steps:\n      - &step\n        name: build\n        exec: tool\n        <<: *step\n"

	if os.Getenv(helperEnv) == "1" {
		_, err := Load(writePipeline(t, contents))
		if err == nil || !strings.Contains(err.Error(), "circular merge key reference") {
			t.Fatalf("Load() error = %v, want circular merge key reference", err)
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestLoad_CircularMergeKeyRejectedWithoutCrash$", "-test.count=1")
	command.Env = append(os.Environ(), helperEnv+"=1")
	output, err := command.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("Load() did not reject circular merge key before timeout: %v\n%s", ctx.Err(), output)
	}
	if err != nil {
		t.Fatalf("circular merge helper failed: %v\n%s", err, output)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.yaml")

	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() error = nil, want file not found error")
	}
	if !strings.Contains(err.Error(), path) {
		t.Errorf("Load() error = %q, want path %q", err, path)
	}
}

func validPipelineFile() PipelineFile {
	return PipelineFile{
		Version: 1,
		Pipelines: map[string]Pipeline{
			"default": {
				Steps: []Step{{Name: "build", Exec: "tool"}},
			},
		},
	}
}

func writePipeline(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "watt.yaml")
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
