package proc

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

type startedProcess struct {
	process    windows.Handle
	thread     windows.Handle
	stdoutRead *os.File
	stderrRead *os.File
	stdoutDone <-chan struct{}
	stderrDone <-chan struct{}
}

// startProcess creates the child suspended (CREATE_SUSPENDED) in a new process
// group (CREATE_NEW_PROCESS_GROUP) so the caller can join it to a Job Object
// before it runs any user code. stdout/stderr are captured through inheritable
// pipes; the read ends are returned and forwarded by the caller.
func startProcess(spec Spec) (*startedProcess, bool, error) {
	stdoutRead, stdoutWrite, err := createPipe()
	if err != nil {
		return nil, false, fmt.Errorf("create stdout pipe: %w", err)
	}
	stderrRead, stderrWrite, err := createPipe()
	if err != nil {
		closePipePair(stdoutRead, stdoutWrite)
		return nil, false, fmt.Errorf("create stderr pipe: %w", err)
	}
	stdinHandle, err := openNull()
	if err != nil {
		closePipePair(stdoutRead, stdoutWrite)
		closePipePair(stderrRead, stderrWrite)
		return nil, false, fmt.Errorf("open null device: %w", err)
	}

	var startupInfo windows.StartupInfo
	startupInfo.Cb = uint32(unsafe.Sizeof(startupInfo))
	startupInfo.Flags = windows.STARTF_USESTDHANDLES
	startupInfo.StdInput = stdinHandle
	startupInfo.StdOutput = stdoutWrite
	startupInfo.StdErr = stderrWrite

	appName, err := windows.UTF16PtrFromString(spec.Path)
	if err != nil {
		closePipePair(stdoutRead, stdoutWrite)
		closePipePair(stderrRead, stderrWrite)
		_ = windows.CloseHandle(stdinHandle)
		return nil, false, fmt.Errorf("encode application path: %w", err)
	}

	commandLine := buildCommandLine(append([]string{spec.Path}, spec.Args...))

	var currentDir *uint16
	if spec.Dir != "" {
		currentDir, err = windows.UTF16PtrFromString(spec.Dir)
		if err != nil {
			closePipePair(stdoutRead, stdoutWrite)
			closePipePair(stderrRead, stderrWrite)
			_ = windows.CloseHandle(stdinHandle)
			return nil, false, fmt.Errorf("encode working directory: %w", err)
		}
	}

	var envBlock *uint16
	if len(spec.Env) > 0 {
		envBlock = buildEnvBlock(spec.Env)
	}

	creationFlags := uint32(windows.CREATE_SUSPENDED | windows.CREATE_NEW_PROCESS_GROUP | windows.CREATE_UNICODE_ENVIRONMENT)

	var processInfo windows.ProcessInformation
	err = windows.CreateProcess(
		appName,
		&commandLine[0],
		nil,
		nil,
		true,
		creationFlags,
		envBlock,
		currentDir,
		&startupInfo,
		&processInfo,
	)
	if err != nil {
		notFound := errors.Is(err, windows.ERROR_FILE_NOT_FOUND) || errors.Is(err, windows.ERROR_PATH_NOT_FOUND)
		closePipePair(stdoutRead, stdoutWrite)
		closePipePair(stderrRead, stderrWrite)
		_ = windows.CloseHandle(stdinHandle)
		return nil, notFound, err
	}

	// The child inherited copies of the write ends and stdin; close ours.
	_ = windows.CloseHandle(stdoutWrite)
	_ = windows.CloseHandle(stderrWrite)
	_ = windows.CloseHandle(stdinHandle)

	return &startedProcess{
		process:    processInfo.Process,
		thread:     processInfo.Thread,
		stdoutRead: stdoutRead,
		stderrRead: stderrRead,
	}, false, nil
}

func (p *startedProcess) forward(stdout, stderr io.Writer) {
	p.stdoutDone = copyStream(stdout, p.stdoutRead)
	p.stderrDone = copyStream(stderr, p.stderrRead)
}

func (p *startedProcess) closeProcessHandles() {
	if p.thread != 0 {
		_ = windows.CloseHandle(p.thread)
		p.thread = 0
	}
	if p.process != 0 {
		_ = windows.CloseHandle(p.process)
		p.process = 0
	}
}

// waitForIO blocks until the stdout/stderr forwarding goroutines finish. It is
// safe only after the child's write ends are closed (the process has exited or
// been terminated); otherwise it would block on open pipe handles.
func (p *startedProcess) waitForIO() {
	if p.stdoutDone != nil {
		<-p.stdoutDone
	}
	if p.stderrDone != nil {
		<-p.stderrDone
	}
}

func (p *startedProcess) closeIO() {
	if p.stdoutRead != nil {
		_ = p.stdoutRead.Close()
		p.stdoutRead = nil
	}
	if p.stderrRead != nil {
		_ = p.stderrRead.Close()
		p.stderrRead = nil
	}
}

func copyStream(dst io.Writer, src *os.File) <-chan struct{} {
	done := make(chan struct{})
	if dst == nil {
		dst = io.Discard
	}
	go func() {
		defer close(done)
		_, _ = io.Copy(dst, src)
	}()
	return done
}

func createPipe() (*os.File, windows.Handle, error) {
	var sa windows.SecurityAttributes
	sa.Length = uint32(unsafe.Sizeof(sa))
	sa.InheritHandle = 1

	var readHandle, writeHandle windows.Handle
	if err := windows.CreatePipe(&readHandle, &writeHandle, &sa, 0); err != nil {
		return nil, 0, err
	}
	// Keep the read end out of the child's handle table so its close is the
	// only thing that ends the stream.
	if err := windows.SetHandleInformation(readHandle, windows.HANDLE_FLAG_INHERIT, 0); err != nil {
		_ = windows.CloseHandle(readHandle)
		_ = windows.CloseHandle(writeHandle)
		return nil, 0, err
	}

	return os.NewFile(uintptr(readHandle), "|0"), writeHandle, nil
}

func closePipePair(read *os.File, write windows.Handle) {
	if read != nil {
		_ = read.Close()
	}
	if write != 0 {
		_ = windows.CloseHandle(write)
	}
}

func openNull() (windows.Handle, error) {
	var sa windows.SecurityAttributes
	sa.Length = uint32(unsafe.Sizeof(sa))
	sa.InheritHandle = 1

	name, err := windows.UTF16PtrFromString("NUL")
	if err != nil {
		return 0, err
	}
	handle, err := windows.CreateFile(
		name,
		windows.GENERIC_READ,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		&sa,
		windows.OPEN_EXISTING,
		windows.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		return 0, err
	}
	return handle, nil
}

// buildCommandLine joins args into a null-terminated UTF-16 command line with
// Windows quoting rules (matching what os/exec would produce).
func buildCommandLine(args []string) []uint16 {
	var builder strings.Builder
	for index, arg := range args {
		if index > 0 {
			builder.WriteByte(' ')
		}
		builder.WriteString(syscall.EscapeArg(arg))
	}
	return windows.StringToUTF16(builder.String())
}

// buildEnvBlock builds a Windows environment block (null-terminated strings,
// terminated by an empty string).
func buildEnvBlock(env []string) *uint16 {
	block := make([]uint16, 0)
	for _, entry := range env {
		for _, r := range entry {
			block = append(block, uint16(r))
		}
		block = append(block, 0)
	}
	block = append(block, 0)
	return &block[0]
}
