package result

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestWritePersistsFrozenResultShape(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".watt", "result.json")
	exitCode := 7
	want := Result{
		SchemaVersion: 1,
		Pipeline:      "default",
		Status:        "failed",
		Steps: []Step{{
			Name:     "build",
			Status:   "failed",
			ExitCode: &exitCode,
			OutputTail: OutputTail{
				Stdout: "out",
				Stderr: "err",
			},
		}},
		Environment: Environment{OS: "windows", Arch: "amd64"},
	}

	if err := Write(path, want); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var got Result
	if err := json.Unmarshal(contents, &got); err != nil {
		t.Fatal(err)
	}
	if got.SchemaVersion != 1 || got.Pipeline != "default" || got.Status != "failed" {
		t.Fatalf("result identity = %#v, want schema 1/default/failed", got)
	}
	if len(got.Steps) != 1 || got.Steps[0].ExitCode == nil || *got.Steps[0].ExitCode != 7 {
		t.Fatalf("steps = %#v, want one step with exit code 7", got.Steps)
	}
	if !reflect.DeepEqual(got.Environment, want.Environment) {
		t.Errorf("environment = %#v, want %#v", got.Environment, want.Environment)
	}
}

func TestMarshalStepWithoutExitCodeUsesNull(t *testing.T) {
	contents, err := json.Marshal(Result{Steps: []Step{{Name: "cwd", Status: "failed"}}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(contents), `"exit_code":null`) {
		t.Fatalf("JSON = %s, want null exit_code", contents)
	}
}
