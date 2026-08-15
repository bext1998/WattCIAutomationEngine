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

// confirmEmpty reports whether the Job Object holds no live processes. A Job
// Object handle becomes signaled only once every process in the job has
// terminated, which is the authoritative "no live processes" signal (P-3). If
// processes remain, it terminates the whole Job Object and waits up to deadline
// for the handle to become signaled.
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
