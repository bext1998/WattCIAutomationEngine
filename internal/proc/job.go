package proc

import (
	"fmt"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

type job struct {
	handle windows.Handle
}

// createJob creates a Job Object with JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE set at
// creation time (P-1), so closing the handle force-terminates any processes
// that remain inside it.
func createJob() (*job, error) {
	handle, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return nil, fmt.Errorf("create job object: %w", err)
	}

	var info windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION
	info.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	if _, err := windows.SetInformationJobObject(
		handle,
		windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&info)),
		uint32(unsafe.Sizeof(info)),
	); err != nil {
		_ = windows.CloseHandle(handle)
		return nil, fmt.Errorf("set job object kill-on-close: %w", err)
	}

	return &job{handle: handle}, nil
}

func (j *job) close() {
	if j.handle != 0 {
		_ = windows.CloseHandle(j.handle)
		j.handle = 0
	}
}

// confirmEmpty reports whether the Job Object holds no live processes.
//
// Implementation note: this waits on the Job Object handle itself, treating
// "handle became signaled" as "no live processes remain". Microsoft's official
// Job Objects documentation describes the signaled state as being set when all
// processes terminate *because an end-of-job time limit was exceeded*, which is
// worded differently from what we rely on here. Empirical testing on Windows 11
// (10.0.26200, including -race and repeated runs) shows the handle is signaled
// exactly when the process count reaches zero, with no spurious signals, but
// this behavior has NOT been verified across other Windows editions (Server
// Core, containers). The documented, portable way to detect
// JOB_OBJECT_MSG_ACTIVE_PROCESS_ZERO is an I/O completion port; that is a
// larger change and is intentionally out of scope for now.
func (j *job) confirmEmpty(deadline time.Duration) (bool, error) {
	signaled, err := j.signaled(0)
	if err != nil {
		return false, err
	}
	if signaled {
		return true, nil
	}

	_ = windows.TerminateJobObject(j.handle, 1)
	return j.signaled(deadline)
}

func (j *job) signaled(timeout time.Duration) (bool, error) {
	event, err := windows.WaitForSingleObject(j.handle, uint32(timeout.Milliseconds()))
	if err != nil {
		return false, fmt.Errorf("wait for job object: %w", err)
	}
	return event == windows.WAIT_OBJECT_0, nil
}
