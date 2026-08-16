// Package proc manages Windows process trees via Job Objects.
//
// It is the only package that speaks Win32 process/JOB APIs directly. Callers
// (internal/runner) pass an already-resolved executable path; proc never
// resolves PATH itself.
package proc
