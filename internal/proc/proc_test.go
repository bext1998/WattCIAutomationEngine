package proc

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"golang.org/x/sys/windows"
)

func TestRunExecutesAndForwardsOutput(t *testing.T) {
	var stdout, stderr bytes.Buffer
	outcome := Run(context.Background(), Spec{
		Path:   os.Args[0],
		Args:   []string{"-test.run=TestProcHelperProcess"},
		Env:    helperEnv("print"),
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if outcome.Status != StatusExited {
		t.Fatalf("status = %q, want %q; err = %v", outcome.Status, StatusExited, outcome.Err)
	}
	if outcome.ExitCode != 0 {
		t.Errorf("exit code = %d, want 0", outcome.ExitCode)
	}
	if stdout.String() != "proc-stdout" {
		t.Errorf("stdout = %q, want %q", stdout.String(), "proc-stdout")
	}
	if stderr.String() != "proc-stderr" {
		t.Errorf("stderr = %q, want %q", stderr.String(), "proc-stderr")
	}
}

func TestRunReportsExitCode(t *testing.T) {
	outcome := Run(context.Background(), Spec{
		Path: os.Args[0],
		Args: []string{"-test.run=TestProcHelperProcess"},
		Env:  helperEnv("exit", "WATT_PROC_EXIT=7"),
	})

	if outcome.Status != StatusExited {
		t.Fatalf("status = %q, want %q; err = %v", outcome.Status, StatusExited, outcome.Err)
	}
	if outcome.ExitCode != 7 {
		t.Errorf("exit code = %d, want 7", outcome.ExitCode)
	}
}

func TestRunStartFailedNotFound(t *testing.T) {
	outcome := Run(context.Background(), Spec{
		Path: filepath.Join(t.TempDir(), "does-not-exist.exe"),
		Env:  helperEnv("print"),
	})

	if outcome.Status != StatusStartFailed {
		t.Fatalf("status = %q, want %q", outcome.Status, StatusStartFailed)
	}
	if !outcome.NotFound {
		t.Error("NotFound = false, want true")
	}
}

func TestProc_NoOrphansOnNormalCompletion(t *testing.T) {
	pidFile := filepath.Join(t.TempDir(), "pids")

	outcome := Run(context.Background(), Spec{
		Path: os.Args[0],
		Args: []string{"-test.run=TestProcHelperProcess"},
		Env:  helperEnv("spawn_sleeper", "WATT_PROC_PIDFILE="+pidFile),
	})

	if outcome.Status != StatusExited {
		t.Fatalf("status = %q, want %q; err = %v", outcome.Status, StatusExited, outcome.Err)
	}
	if outcome.ExitCode != 0 {
		t.Errorf("exit code = %d, want 0", outcome.ExitCode)
	}

	_, grandchild := readPIDs(t, pidFile)
	if processAlive(grandchild) {
		t.Errorf("grandchild pid %d still alive, want cleaned up", grandchild)
	}
}

func TestCancel_KillsDescendantProcesses(t *testing.T) {
	pidFile := filepath.Join(t.TempDir(), "pids")
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan Outcome, 1)
	go func() {
		done <- Run(ctx, Spec{
			Path: os.Args[0],
			Args: []string{"-test.run=TestProcHelperProcess"},
			Env:  helperEnv("spawn_sleeper_and_sleep", "WATT_PROC_PIDFILE="+pidFile),
		})
	}()

	helper, grandchild := readPIDs(t, pidFile)
	cancel()

	outcome := <-done
	if outcome.Status != StatusCancelled {
		t.Fatalf("status = %q, want %q; err = %v", outcome.Status, StatusCancelled, outcome.Err)
	}
	if processAlive(helper) {
		t.Errorf("helper pid %d still alive, want terminated", helper)
	}
	if processAlive(grandchild) {
		t.Errorf("grandchild pid %d still alive, want terminated", grandchild)
	}
}

func TestExitCode_InternalErrorOnUnconfirmedCancellation(t *testing.T) {
	pidFile := filepath.Join(t.TempDir(), "pids")
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan Outcome, 1)
	go func() {
		done <- Run(ctx, Spec{
			Path: os.Args[0],
			Args: []string{"-test.run=TestProcHelperProcess"},
			Env:  helperEnv("spawn_sleeper_and_sleep", "WATT_PROC_PIDFILE="+pidFile),
			// A deadline too short for even an instant termination to be
			// confirmed, forcing the close-on-kill fallback.
			ConfirmDeadline: time.Nanosecond,
		})
	}()

	helper, grandchild := readPIDs(t, pidFile)
	cancel()

	outcome := <-done
	if outcome.Status != StatusInternalError {
		t.Fatalf("status = %q, want %q", outcome.Status, StatusInternalError)
	}
	waitForProcessGone(t, helper)
	waitForProcessGone(t, grandchild)
}

// --- helpers ---

func helperEnv(mode string, extra ...string) []string {
	env := append([]string{}, os.Environ()...)
	env = append(env, "WATT_PROC_HELPER=1", "WATT_PROC_MODE="+mode)
	env = append(env, extra...)
	return env
}

func readPIDs(t *testing.T, path string) (int, int) {
	t.Helper()
	waitFor(t, 10*time.Second, func() bool {
		_, err := os.Stat(path)
		return err == nil
	})
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read pid file: %v", err)
	}
	var helper, grandchild int
	if _, err := fmt.Sscanf(string(contents), "%d %d", &helper, &grandchild); err != nil {
		t.Fatalf("parse pid file %q: %v", contents, err)
	}
	return helper, grandchild
}

func processAlive(pid int) bool {
	handle, err := windows.OpenProcess(windows.SYNCHRONIZE, false, uint32(pid))
	if err != nil {
		return false
	}
	defer windows.CloseHandle(handle)
	event, _ := windows.WaitForSingleObject(handle, 0)
	return event == uint32(windows.WAIT_TIMEOUT)
}

func waitForProcessGone(t *testing.T, pid int) {
	t.Helper()
	waitFor(t, 10*time.Second, func() bool {
		return !processAlive(pid)
	})
}

func waitFor(t *testing.T, timeout time.Duration, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition not met within timeout")
}

func TestProcHelperProcess(t *testing.T) {
	if os.Getenv("WATT_PROC_HELPER") != "1" {
		return
	}
	switch os.Getenv("WATT_PROC_MODE") {
	case "print":
		fmt.Fprint(os.Stdout, "proc-stdout")
		fmt.Fprint(os.Stderr, "proc-stderr")
		os.Exit(0)
	case "exit":
		code, _ := strconv.Atoi(os.Getenv("WATT_PROC_EXIT"))
		os.Exit(code)
	case "sleeper":
		time.Sleep(60 * time.Second)
		os.Exit(0)
	case "spawn_sleeper":
		grandchild := spawnSleeper()
		writePIDs(os.Getpid(), grandchild)
		os.Exit(0)
	case "spawn_sleeper_and_sleep":
		grandchild := spawnSleeper()
		writePIDs(os.Getpid(), grandchild)
		time.Sleep(60 * time.Second)
		os.Exit(0)
	}
}

func spawnSleeper() int {
	_ = os.Setenv("WATT_PROC_MODE", "sleeper")
	command := exec.Command(os.Args[0], "-test.run=TestProcHelperProcess")
	if err := command.Start(); err != nil {
		os.Exit(2)
	}
	return command.Process.Pid
}

func writePIDs(helper, grandchild int) {
	pidFile := os.Getenv("WATT_PROC_PIDFILE")
	if pidFile == "" {
		return
	}
	_ = os.WriteFile(pidFile, []byte(fmt.Sprintf("%d %d", helper, grandchild)), 0o644)
}
