package proc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"golang.org/x/sys/windows"
)

// Status is the outcome classification of a single proc.Run invocation.
type Status string

const (
	// StatusExited means the process ran to completion; ExitCode is valid.
	StatusExited Status = "exited"
	// StatusStartFailed means the process could not be started.
	StatusStartFailed Status = "start_failed"
	// StatusCancelled means ctx was cancelled and the whole process tree was
	// confirmed terminated within the deadline.
	StatusCancelled Status = "cancelled"
	// StatusInternalError means Watt could not confirm the Job Object became
	// empty within the deadline (close-on-kill was used as the fallback), or an
	// infrastructure step (job creation, assignment, wait) failed.
	StatusInternalError Status = "internal_error"
)

// Spec describes the process to start. Path must already be resolved to a real
// executable (proc does not search PATH).
type Spec struct {
	Path   string
	Args   []string
	Env    []string // "KEY=VALUE" entries, in the exact form to pass to the child
	Dir    string   // working directory; empty means inherit Watt's
	Stdout io.Writer
	Stderr io.Writer

	// ConfirmDeadline is how long to wait for the Job Object to become empty
	// after a termination signal before falling back to close-on-kill.
	// A non-positive value uses the default (5 seconds, A-10). Production
	// callers must leave it unset; tests may shrink it.
	ConfirmDeadline time.Duration
}

// Outcome is the result of a proc.Run invocation.
type Outcome struct {
	Status   Status
	ExitCode int  // valid only when Status == StatusExited
	NotFound bool // valid only when Status == StatusStartFailed
	Err      error
}

const defaultConfirmDeadline = 5 * time.Second

// Run starts Path inside a fresh Windows Job Object and waits for it to exit
// or for ctx to be cancelled.
//
// The Job Object is created with JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE before the
// process is resumed (P-1). The process is created suspended and joined to the
// Job Object before any user code runs. On cancellation, the whole Job Object
// is terminated (P-2). Before the Job Object handle is closed, Run confirms it
// holds no live processes (P-3); if the deadline elapses first, the handle is
// closed and the OS kill-on-close guarantee becomes the last-resort cleanup.
func Run(ctx context.Context, spec Spec) Outcome {
	if ctx == nil {
		ctx = context.Background()
	}
	deadline := spec.ConfirmDeadline
	if deadline <= 0 {
		deadline = defaultConfirmDeadline
	}

	job, err := createJob()
	if err != nil {
		return Outcome{Status: StatusInternalError, Err: err}
	}

	started, notFound, err := startProcess(spec)
	if err != nil {
		job.close()
		return Outcome{Status: StatusStartFailed, NotFound: notFound, Err: err}
	}

	if err := windows.AssignProcessToJobObject(job.handle, started.process); err != nil {
		_ = windows.TerminateProcess(started.process, 1)
		started.closeProcessHandles()
		started.closeIO()
		job.close()
		return Outcome{Status: StatusInternalError, Err: fmt.Errorf("assign process to job object: %w", err)}
	}

	started.forward(spec.Stdout, spec.Stderr)

	if _, err := windows.ResumeThread(started.thread); err != nil {
		_ = windows.TerminateJobObject(job.handle, 1)
		started.closeProcessHandles()
		started.waitForIO()
		started.closeIO()
		job.close()
		return Outcome{Status: StatusInternalError, Err: fmt.Errorf("resume process: %w", err)}
	}

	cancelled := false
	exitCode := 0
	for {
		if ctx.Err() != nil {
			cancelled = true
			_ = windows.TerminateJobObject(job.handle, 1)
			break
		}

		event, waitErr := windows.WaitForSingleObject(started.process, 10)
		if waitErr != nil {
			_ = windows.TerminateJobObject(job.handle, 1)
			started.closeProcessHandles()
			started.waitForIO()
			started.closeIO()
			job.close()
			return Outcome{Status: StatusInternalError, Err: fmt.Errorf("wait for process: %w", waitErr)}
		}
		if event == windows.WAIT_OBJECT_0 {
			var code uint32
			if err := windows.GetExitCodeProcess(started.process, &code); err != nil {
				started.closeProcessHandles()
				started.waitForIO()
				started.closeIO()
				job.close()
				return Outcome{Status: StatusInternalError, Err: fmt.Errorf("get exit code: %w", err)}
			}
			exitCode = int(code)
			break
		}
	}

	started.closeProcessHandles()

	confirmed, confirmErr := job.confirmEmpty(deadline)
	// Closing the handle (with KILL_ON_JOB_CLOSE already set) is the last-resort
	// cleanup when confirmEmpty could not reach zero within the deadline.
	job.close()

	started.waitForIO()
	started.closeIO()

	if confirmErr != nil {
		return Outcome{Status: StatusInternalError, Err: confirmErr}
	}
	if !confirmed {
		return Outcome{Status: StatusInternalError, Err: errors.New("job object did not become empty within the confirmation deadline")}
	}
	if cancelled {
		return Outcome{Status: StatusCancelled}
	}
	return Outcome{Status: StatusExited, ExitCode: exitCode}
}
