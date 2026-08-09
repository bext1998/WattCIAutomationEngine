package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	path := writePipeline(t, "version: 1\nunknown: true\npipelines: {}\n")

	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() error = nil, want unknown field error")
	}
	if !strings.Contains(err.Error(), "unknown") {
		t.Errorf("Load() error = %q, want unknown field context", err)
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
