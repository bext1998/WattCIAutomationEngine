package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/bext1998/WattCIAutomationEngine/internal/result"
)

func TestRootCommandVersionOutput(t *testing.T) {
	command := newRootCommand()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.SetOut(&stdout)
	command.SetErr(&stderr)
	command.SetArgs([]string{"--version"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if got, want := stdout.String(), "watt dev\n"; got != want {
		t.Errorf("version output = %q, want %q", got, want)
	}
	if got := stderr.String(); got != "" {
		t.Errorf("version stderr = %q, want empty", got)
	}
}

func TestRootCommandRunReportsInvalidPipeline(t *testing.T) {
	command := newRootCommand()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.SetOut(&stdout)
	command.SetErr(&stderr)
	command.SetArgs([]string{"run"})

	if got, want := execute(command), EXIT_INVALID_PIPELINE; got != want {
		t.Errorf("exit code = %d, want %d", got, want)
	}
	if got := stdout.String(); got != "" {
		t.Errorf("stdout = %q, want empty", got)
	}
	if got := stderr.String(); !strings.Contains(got, "watt.yaml") {
		t.Errorf("stderr = %q, want pipeline path", got)
	}
}

func TestRun_DefaultPipelineSelection(t *testing.T) {
	code, _, stderr, got := runFixture(t, "", pipelineWithSteps(helperStep("run", 0)))

	if code != EXIT_SUCCESS {
		t.Fatalf("exit code = %d, want %d; stderr = %q", code, EXIT_SUCCESS, stderr)
	}
	if got.Pipeline != "default" || got.Status != "success" || len(got.Steps) != 1 {
		t.Fatalf("result = %#v, want default success with one step", got)
	}
}

func TestRun_NamedPipelineSelection(t *testing.T) {
	contents := "version: 1\npipelines:\n  default:\n    steps:\n" + indent(helperStep("default", 9), 6) +
		"  package:\n    steps:\n" + indent(helperStep("package", 0), 6)
	code, _, stderr, got := runFixture(t, "package", contents)

	if code != EXIT_SUCCESS {
		t.Fatalf("exit code = %d, want %d; stderr = %q", code, EXIT_SUCCESS, stderr)
	}
	if got.Pipeline != "package" || got.Status != "success" {
		t.Fatalf("result identity = %#v, want package success", got)
	}
}

func TestRun_FailFastStopsAtFailedStep(t *testing.T) {
	contents := pipelineWithSteps(
		helperStep("first", 0),
		helperStep("second", 9),
		helperStep("third", 0),
	)
	code, _, stderr, got := runFixture(t, "", contents)

	if code != EXIT_STEP_FAILED {
		t.Fatalf("exit code = %d, want %d; stderr = %q", code, EXIT_STEP_FAILED, stderr)
	}
	if got.Status != "failed" || len(got.Steps) != 2 {
		t.Fatalf("result = %#v, want failed result with two steps", got)
	}
	if got.Steps[1].ExitCode == nil || *got.Steps[1].ExitCode != 9 {
		t.Fatalf("failed step exit code = %v, want 9", got.Steps[1].ExitCode)
	}
}

func TestResult_PartialOnFailure(t *testing.T) {
	contents := pipelineWithSteps(
		helperStep("first", 0),
		helperStep("second", 9),
		helperStep("not-started", 0),
	)
	code, _, _, got := runFixture(t, "", contents)

	if code != EXIT_STEP_FAILED || len(got.Steps) != 2 {
		t.Fatalf("code = %d, steps = %d, want step failure with two attempted steps", code, len(got.Steps))
	}
}

func TestExecStep_NoShellIndirection(t *testing.T) {
	argvPath := filepath.Join(t.TempDir(), "argv")
	helper := buildArgvHelper(t)
	args := []string{"&", "|", ">", "$(WATT_NOT_A_COMMAND)", "contains spaces"}
	var argumentLines strings.Builder
	for _, arg := range args {
		fmt.Fprintf(&argumentLines, "          - '%s'\n", yamlSingleQuote(arg))
	}
	pipeline := fmt.Sprintf("version: 1\npipelines:\n  default:\n    steps:\n      - name: direct\n        exec: '%s'\n        args:\n%s        env:\n          WATT_ARGV_PATH: '%s'\n", yamlSingleQuote(helper), argumentLines.String(), yamlSingleQuote(argvPath))

	code, _, stderr, got := runFixture(t, "", pipeline)
	if code != EXIT_SUCCESS || got.Status != "success" {
		t.Fatalf("code = %d, status = %q, stderr = %q", code, got.Status, stderr)
	}

	contents, err := os.ReadFile(argvPath)
	if err != nil {
		t.Fatal(err)
	}
	gotArgs := strings.Split(string(contents), "\x00")
	if len(gotArgs) == 0 {
		t.Fatal("helper received no arguments")
	}
	if !reflect.DeepEqual(gotArgs[1:], args) {
		t.Errorf("helper args = %#v, want %#v", gotArgs[1:], args)
	}
}

func TestOutputPassthrough_RealtimeStdoutStderr(t *testing.T) {
	code, stdout, stderr, got := runFixture(t, "", pipelineWithSteps(helperStep("output", 0)))

	if code != EXIT_SUCCESS || got.Steps[0].OutputTail.Stdout != "stdout" || got.Steps[0].OutputTail.Stderr != "stderr" {
		t.Fatalf("code = %d, stdout=%q, stderr=%q, result tail=%#v", code, stdout, stderr, got.Steps[0].OutputTail)
	}
}

func TestResult_ExitCodeNullOnEnvironmentUnavailable(t *testing.T) {
	contents := pipelineWithSteps("      - name: missing\n        exec: watt-command-that-does-not-exist\n")
	code, _, stderr, got := runFixture(t, "", contents)

	if code != EXIT_ENVIRONMENT_UNAVAILABLE || got.Status != "environment_unavailable" {
		t.Fatalf("code = %d, status = %q, stderr = %q", code, got.Status, stderr)
	}
	if got.Steps[0].ExitCode != nil || got.Steps[0].OutputTail.Stdout != "" || got.Steps[0].OutputTail.Stderr != "" {
		t.Fatalf("step = %#v, want null exit code and empty output", got.Steps[0])
	}
	if !strings.Contains(stderr, "watt-command-that-does-not-exist") {
		t.Errorf("stderr = %q, want missing command name", stderr)
	}
}

func TestRun_FailFastStopsAfterCwdFailure(t *testing.T) {
	missingCwd := filepath.Join(t.TempDir(), "missing")
	thirdStepMarker := filepath.Join(t.TempDir(), "third-step-argv")
	contents := pipelineWithSteps(
		helperStep("first", 0),
		fmt.Sprintf("      - name: cwd\n        exec: '%s'\n        args:\n          - -test.run=TestRunHelperProcess\n        env:\n          WATT_RUN_HELPER: '1'\n        cwd: '%s'\n", yamlSingleQuote(os.Args[0]), yamlSingleQuote(missingCwd)),
		fmt.Sprintf("      - name: third\n        exec: '%s'\n        args:\n          - -test.run=TestRunHelperProcess\n        env:\n          WATT_RUN_HELPER: '1'\n          WATT_RUN_ARGV_PATH: '%s'\n", yamlSingleQuote(os.Args[0]), yamlSingleQuote(thirdStepMarker)),
	)

	code, _, stderr, got := runFixture(t, "", contents)
	if code != EXIT_STEP_FAILED || got.Status != "failed" {
		t.Fatalf("code = %d, status = %q, stderr = %q", code, got.Status, stderr)
	}
	if len(got.Steps) != 2 || got.Steps[0].Status != "success" || got.Steps[1].Status != "failed" {
		t.Fatalf("steps = %#v, want successful first and failed cwd step only", got.Steps)
	}
	if got.Steps[1].ExitCode != nil || got.Steps[1].OutputTail != (result.OutputTail{}) {
		t.Errorf("cwd step = %#v, want null exit code and empty output", got.Steps[1])
	}
	if _, err := os.Stat(thirdStepMarker); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("third step marker exists or could not be checked: %v", err)
	}
}

func TestRun_ResultWriteFailureReturnsInternalError(t *testing.T) {
	workdir := t.TempDir()
	if err := os.WriteFile(filepath.Join(workdir, "watt.yaml"), []byte(pipelineWithSteps(helperStep("success", 0))), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workdir, ".watt"), []byte("blocks result directory"), 0o644); err != nil {
		t.Fatal(err)
	}
	oldWorkingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(workdir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldWorkingDirectory); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	command := newRootCommand()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.SetOut(&stdout)
	command.SetErr(&stderr)
	command.SetArgs([]string{"run"})
	if got := execute(command); got != EXIT_INTERNAL_ERROR {
		t.Fatalf("exit code = %d, want %d; stderr = %q", got, EXIT_INTERNAL_ERROR, stderr.String())
	}
	if !strings.Contains(stderr.String(), "create result directory") {
		t.Errorf("stderr = %q, want result write failure", stderr.String())
	}
}

func TestResult_ExitCodeNullOnCwdFailure(t *testing.T) {
	missingCwd := filepath.Join(t.TempDir(), "missing")
	contents := pipelineWithSteps(fmt.Sprintf("      - name: cwd\n        exec: '%s'\n        args:\n          - -test.run=TestRunHelperProcess\n        env:\n          WATT_RUN_HELPER: '1'\n        cwd: '%s'\n", yamlSingleQuote(os.Args[0]), yamlSingleQuote(missingCwd)))
	code, _, stderr, got := runFixture(t, "", contents)

	if code != EXIT_STEP_FAILED || got.Status != "failed" {
		t.Fatalf("code = %d, status = %q, stderr = %q", code, got.Status, stderr)
	}
	if got.Steps[0].ExitCode != nil || got.Steps[0].OutputTail != (result.OutputTail{}) {
		t.Fatalf("step = %#v, want null exit code and empty output", got.Steps[0])
	}
}

func runFixture(t *testing.T, pipelineName, contents string) (int, string, string, result.Result) {
	t.Helper()
	workdir := t.TempDir()
	if err := os.WriteFile(filepath.Join(workdir, "watt.yaml"), []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
	oldWorkingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(workdir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldWorkingDirectory); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	command := newRootCommand()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.SetOut(&stdout)
	command.SetErr(&stderr)
	args := []string{"run"}
	if pipelineName != "" {
		args = append(args, pipelineName)
	}
	command.SetArgs(args)
	code := execute(command)

	resultContents, err := os.ReadFile(filepath.Join(workdir, ".watt", "result.json"))
	if err != nil {
		t.Fatalf("read result: %v; stdout=%q stderr=%q", err, stdout.String(), stderr.String())
	}
	var got result.Result
	if err := json.Unmarshal(resultContents, &got); err != nil {
		t.Fatalf("decode result: %v", err)
	}
	return code, stdout.String(), stderr.String(), got
}

func buildArgvHelper(t *testing.T) string {
	t.Helper()
	directory := t.TempDir()
	source := filepath.Join(directory, "main.go")
	binary := filepath.Join(directory, "argv-helper.exe")
	contents := `package main

import (
	"os"
	"strings"
)

func main() {
	_ = os.WriteFile(os.Getenv("WATT_ARGV_PATH"), []byte(strings.Join(os.Args, "\x00")), 0o644)
}
`
	if err := os.WriteFile(source, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command("go", "build", "-o", binary, source).CombinedOutput(); err != nil {
		t.Fatalf("build argv helper: %v\n%s", err, output)
	}
	return binary
}

func pipelineWithSteps(steps ...string) string {
	return "version: 1\npipelines:\n  default:\n    steps:\n" + strings.Join(steps, "")
}

func helperStep(name string, exitCode int) string {
	return fmt.Sprintf("      - name: %s\n        exec: '%s'\n        args:\n          - -test.run=TestRunHelperProcess\n        env:\n          WATT_RUN_HELPER: '1'\n          WATT_RUN_EXIT: '%d'\n", name, yamlSingleQuote(os.Args[0]), exitCode)
}

func yamlSingleQuote(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}

func indent(value string, spaces int) string {
	prefix := strings.Repeat(" ", spaces)
	return strings.ReplaceAll(value, "      ", prefix)
}

func TestRunHelperProcess(t *testing.T) {
	if os.Getenv("WATT_RUN_HELPER") != "1" {
		return
	}
	if argvPath := os.Getenv("WATT_RUN_ARGV_PATH"); argvPath != "" {
		_ = os.WriteFile(argvPath, []byte(strings.Join(os.Args, "\x00")), 0o644)
	}
	fmt.Fprint(os.Stdout, "stdout")
	fmt.Fprint(os.Stderr, "stderr")
	exitCode, err := strconv.Atoi(os.Getenv("WATT_RUN_EXIT"))
	if err != nil {
		exitCode = 0
	}
	os.Exit(exitCode)
}

func TestCheck_NoSideEffects(t *testing.T) {
	workdir := t.TempDir()
	oldWorkingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(workdir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldWorkingDirectory); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	resultDirectory := filepath.Join(workdir, ".watt")
	if err := os.Mkdir(resultDirectory, 0o755); err != nil {
		t.Fatal(err)
	}
	resultPath := filepath.Join(resultDirectory, "result.json")
	resultContents := []byte(`{"status":"existing"}`)
	if err := os.WriteFile(resultPath, resultContents, 0o644); err != nil {
		t.Fatal(err)
	}
	pipelineContents := []byte("version: 1\npipelines:\n  default:\n    steps:\n      - name: destructive\n        exec: cmd.exe\n        args:\n          - /c\n          - echo started>started.marker\n")
	if err := os.WriteFile(filepath.Join(workdir, "watt.yaml"), pipelineContents, 0o644); err != nil {
		t.Fatal(err)
	}

	before := snapshotFiles(t, workdir)
	command := newRootCommand()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.SetOut(&stdout)
	command.SetErr(&stderr)
	command.SetArgs([]string{"check"})

	if got, want := execute(command), EXIT_SUCCESS; got != want {
		t.Fatalf("exit code = %d, want %d; stderr = %q", got, want, stderr.String())
	}
	if got := stdout.String(); got != "" {
		t.Errorf("stdout = %q, want empty", got)
	}
	if got := stderr.String(); got != "" {
		t.Errorf("stderr = %q, want empty", got)
	}

	after := snapshotFiles(t, workdir)
	if !reflect.DeepEqual(after, before) {
		t.Errorf("check changed the filesystem: before=%v after=%v", before, after)
	}
}

func TestCheckEnv_DetectsMissingShell(t *testing.T) {
	workdir := t.TempDir()
	oldWorkingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(workdir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldWorkingDirectory); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})
	t.Setenv("PATH", t.TempDir())

	pipelineContents := []byte("version: 1\npipelines:\n  default:\n    steps:\n      - name: script\n        run: echo hello\n")
	if err := os.WriteFile(filepath.Join(workdir, "watt.yaml"), pipelineContents, 0o644); err != nil {
		t.Fatal(err)
	}

	command := newRootCommand()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.SetOut(&stdout)
	command.SetErr(&stderr)
	command.SetArgs([]string{"check", "--env"})

	if got, want := execute(command), EXIT_ENVIRONMENT_UNAVAILABLE; got != want {
		t.Fatalf("exit code = %d, want %d; stderr = %q", got, want, stderr.String())
	}
	if !strings.Contains(stderr.String(), `missing shell "pwsh"`) {
		t.Errorf("stderr = %q, want missing pwsh shell", stderr.String())
	}
	if got := stdout.String(); got != "" {
		t.Errorf("stdout = %q, want empty", got)
	}
}

func TestCheckEnv_DetectsMissingCommand(t *testing.T) {
	workdir := t.TempDir()
	oldWorkingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(workdir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldWorkingDirectory); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})
	t.Setenv("PATH", t.TempDir())

	const missingCommand = "watt-check-missing-command"
	pipelineContents := []byte("version: 1\npipelines:\n  default:\n    steps:\n      - name: command\n        exec: " + missingCommand + "\n")
	if err := os.WriteFile(filepath.Join(workdir, "watt.yaml"), pipelineContents, 0o644); err != nil {
		t.Fatal(err)
	}

	command := newRootCommand()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.SetOut(&stdout)
	command.SetErr(&stderr)
	command.SetArgs([]string{"check", "--env"})

	if got, want := execute(command), EXIT_ENVIRONMENT_UNAVAILABLE; got != want {
		t.Fatalf("exit code = %d, want %d; stderr = %q", got, want, stderr.String())
	}
	if !strings.Contains(stderr.String(), `missing command "`+missingCommand+`"`) {
		t.Errorf("stderr = %q, want missing command %q", stderr.String(), missingCommand)
	}
	if got := stdout.String(); got != "" {
		t.Errorf("stdout = %q, want empty", got)
	}
}

func TestCheckEnv_ReportsMissingRequirementsInStableOrder(t *testing.T) {
	workdir := t.TempDir()
	oldWorkingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(workdir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldWorkingDirectory); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})
	t.Setenv("PATH", t.TempDir())

	pipelineContents := []byte("version: 1\npipelines:\n  zeta:\n    steps:\n      - name: later-command\n        exec: watt-check-missing-command-2\n  alpha:\n    steps:\n      - name: first-shell\n        run: echo hello\n      - name: first-command\n        exec: watt-check-missing-command\n      - name: duplicate-command\n        exec: watt-check-missing-command\n")
	if err := os.WriteFile(filepath.Join(workdir, "watt.yaml"), pipelineContents, 0o644); err != nil {
		t.Fatal(err)
	}

	command := newRootCommand()
	var stderr bytes.Buffer
	command.SetErr(&stderr)
	command.SetArgs([]string{"check", "--env"})

	if got, want := execute(command), EXIT_ENVIRONMENT_UNAVAILABLE; got != want {
		t.Fatalf("exit code = %d, want %d; stderr = %q", got, want, stderr.String())
	}
	want := `environment unavailable: missing shell "pwsh"; missing command "watt-check-missing-command"; missing command "watt-check-missing-command-2"` + "\n"
	if got := stderr.String(); got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

func TestCheckEnv_SucceedsForCmd(t *testing.T) {
	workdir := t.TempDir()
	oldWorkingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(workdir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldWorkingDirectory); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})
	systemRoot := os.Getenv("SystemRoot")
	if systemRoot == "" {
		t.Fatal("SystemRoot is empty")
	}
	t.Setenv("PATH", filepath.Join(systemRoot, "System32"))

	pipelineContents := []byte("version: 1\npipelines:\n  default:\n    steps:\n      - name: shell\n        run: echo hello\n        shell: cmd\n      - name: command\n        exec: cmd.exe\n")
	if err := os.WriteFile(filepath.Join(workdir, "watt.yaml"), pipelineContents, 0o644); err != nil {
		t.Fatal(err)
	}

	command := newRootCommand()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.SetOut(&stdout)
	command.SetErr(&stderr)
	command.SetArgs([]string{"check", "--env"})

	if got, want := execute(command), EXIT_SUCCESS; got != want {
		t.Fatalf("exit code = %d, want %d; stderr = %q", got, want, stderr.String())
	}
	if got := stdout.String(); got != "" {
		t.Errorf("stdout = %q, want empty", got)
	}
	if got := stderr.String(); got != "" {
		t.Errorf("stderr = %q, want empty", got)
	}
}

func TestCheckEnv_NoSideEffects(t *testing.T) {
	workdir := t.TempDir()
	oldWorkingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(workdir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldWorkingDirectory); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})
	systemRoot := os.Getenv("SystemRoot")
	if systemRoot == "" {
		t.Fatal("SystemRoot is empty")
	}
	t.Setenv("PATH", filepath.Join(systemRoot, "System32"))

	resultDirectory := filepath.Join(workdir, ".watt")
	if err := os.Mkdir(resultDirectory, 0o755); err != nil {
		t.Fatal(err)
	}
	resultPath := filepath.Join(resultDirectory, "result.json")
	resultContents := []byte(`{"status":"existing"}`)
	if err := os.WriteFile(resultPath, resultContents, 0o644); err != nil {
		t.Fatal(err)
	}
	pipelineContents := []byte("version: 1\npipelines:\n  default:\n    steps:\n      - name: destructive\n        exec: cmd.exe\n        args:\n          - /c\n          - echo started>started.marker\n")
	if err := os.WriteFile(filepath.Join(workdir, "watt.yaml"), pipelineContents, 0o644); err != nil {
		t.Fatal(err)
	}

	before := snapshotFiles(t, workdir)
	command := newRootCommand()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.SetOut(&stdout)
	command.SetErr(&stderr)
	command.SetArgs([]string{"check", "--env"})

	if got, want := execute(command), EXIT_SUCCESS; got != want {
		t.Fatalf("exit code = %d, want %d; stderr = %q", got, want, stderr.String())
	}
	if got := stdout.String(); got != "" {
		t.Errorf("stdout = %q, want empty", got)
	}
	if got := stderr.String(); got != "" {
		t.Errorf("stderr = %q, want empty", got)
	}

	after := snapshotFiles(t, workdir)
	if !reflect.DeepEqual(after, before) {
		t.Errorf("check --env changed the filesystem: before=%v after=%v", before, after)
	}
}

func TestCheck_FailurePaths(t *testing.T) {
	tests := []struct {
		name     string
		pipeline string
	}{
		{
			name: "load failure",
		},
		{
			name:     "validation failure",
			pipeline: "version: 1\npipelines:\n  default:\n    steps:\n      - name: invalid\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			workdir := t.TempDir()
			oldWorkingDirectory, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Chdir(workdir); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := os.Chdir(oldWorkingDirectory); err != nil {
					t.Errorf("restore working directory: %v", err)
				}
			})

			if test.pipeline != "" {
				if err := os.WriteFile(filepath.Join(workdir, "watt.yaml"), []byte(test.pipeline), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			command := newRootCommand()
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			command.SetOut(&stdout)
			command.SetErr(&stderr)
			command.SetArgs([]string{"check"})

			if got, want := execute(command), EXIT_INVALID_PIPELINE; got != want {
				t.Errorf("exit code = %d, want %d", got, want)
			}
			if got := stdout.String(); got != "" {
				t.Errorf("stdout = %q, want empty", got)
			}
			if got := stderr.String(); got == "" {
				t.Error("stderr is empty, want an error message")
			}
		})
	}
}

type snapshotEntry struct {
	directory bool
	contents  []byte
}

func snapshotFiles(t *testing.T, root string) map[string]snapshotEntry {
	t.Helper()
	snapshot := make(map[string]snapshotEntry)
	if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relativePath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			snapshot[relativePath] = snapshotEntry{directory: true}
			return nil
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		snapshot[relativePath] = snapshotEntry{contents: contents}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return snapshot
}

func TestRootCommandUsageErrorsShareNonRetryableInputExitCode(t *testing.T) {
	if EXIT_USAGE != EXIT_INVALID_PIPELINE {
		t.Fatalf("EXIT_USAGE = %d, want EXIT_INVALID_PIPELINE (%d)", EXIT_USAGE, EXIT_INVALID_PIPELINE)
	}

	tests := []struct {
		name         string
		args         []string
		errorMessage string
	}{
		{name: "unknown command", args: []string{"bogus"}, errorMessage: `unknown command "bogus"`},
		{name: "unknown flag", args: []string{"--bogus"}, errorMessage: "unknown flag: --bogus"},
		{name: "run unexpected arguments", args: []string{"run", "foo", "bar"}, errorMessage: "run accepts at most one pipeline name"},
		{name: "check unexpected arguments", args: []string{"check", "foo"}, errorMessage: `unknown command "foo"`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command := newRootCommand()
			var stderr bytes.Buffer
			command.SetErr(&stderr)
			command.SetArgs(test.args)

			if got := execute(command); got != EXIT_USAGE {
				t.Errorf("exit code = %d, want EXIT_USAGE (%d)", got, EXIT_USAGE)
			}
			if got := strings.Count(stderr.String(), test.errorMessage); got != 1 {
				t.Errorf("stderr contains %q %d times, want exactly once: %q", test.errorMessage, got, stderr.String())
			}
		})
	}
}

func TestExitErrorPreservesCodeAndCause(t *testing.T) {
	cause := errors.New("underlying failure")
	err := &exitError{code: EXIT_INTERNAL_ERROR, err: cause}

	if got, want := err.Error(), cause.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
	if !errors.Is(err, cause) {
		t.Error("errors.Is() did not find the wrapped cause")
	}
}

func TestExecuteUnclassifiedErrorUsesInternalErrorClass(t *testing.T) {
	command := &cobra.Command{
		Use: "watt",
		RunE: func(*cobra.Command, []string) error {
			return errors.New("unexpected failure")
		},
	}
	var stderr bytes.Buffer
	command.SetErr(&stderr)

	if got, want := execute(command), EXIT_INTERNAL_ERROR; got != want {
		t.Errorf("unclassified error exit code = %d, want EXIT_INTERNAL_ERROR (%d)", got, want)
	}
	if !strings.Contains(stderr.String(), "unexpected failure") {
		t.Errorf("stderr = %q, want the unclassified error", stderr.String())
	}
}

func TestRootCommandHelpExcludesCompletion(t *testing.T) {
	command := newRootCommand()
	var stdout bytes.Buffer
	command.SetOut(&stdout)
	command.SetArgs([]string{"--help"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	help := stdout.String()
	for _, name := range []string{"check", "run"} {
		if !strings.Contains(help, "  "+name) {
			t.Errorf("help output does not list %q: %q", name, help)
		}
	}
	if strings.Contains(help, "  completion") {
		t.Errorf("help output exposes completion command: %q", help)
	}
}
