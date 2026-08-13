package runner

import (
	"bytes"
	"fmt"
	"os"
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
