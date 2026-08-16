package env

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

const redactedValue = "[REDACTED]"

var pwshVersionProbeTimeout = 3 * time.Second

// Diagnostics is the plain-data shape of the result.json "environment" block,
// kept independent of the result package so env stays a dependency-free domain
// module (see spec §6.3).
type Diagnostics struct {
	OS             string
	Arch           string
	ShellAvailable map[string]any
	ResolvedTools  map[string]string
	EnvVarNames    []string
}

// ProbeDiagnostics probes shell availability (fixed pwsh/cmd/bash set) and
// resolves the given tool names, returning safe diagnostics that never include
// environment variable values (spec R-4).
func ProbeDiagnostics(osName, arch string, tools []string) Diagnostics {
	diagnostics := Diagnostics{
		OS:             osName,
		Arch:           arch,
		ShellAvailable: make(map[string]any, 3),
		ResolvedTools:  make(map[string]string),
		EnvVarNames:    environmentVariableNames(),
	}

	for _, shell := range []string{"pwsh", "cmd", "bash"} {
		path, err := ResolveExecutable(shell)
		if err != nil {
			diagnostics.ShellAvailable[shell] = false
			continue
		}
		if shell != "pwsh" {
			diagnostics.ShellAvailable[shell] = true
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), pwshVersionProbeTimeout)
		version, err := exec.CommandContext(ctx, path, "-NoLogo", "-NoProfile", "-Command", "$PSVersionTable.PSVersion.ToString()").Output()
		cancel()
		if err != nil || strings.TrimSpace(string(version)) == "" {
			diagnostics.ShellAvailable[shell] = true
			continue
		}
		diagnostics.ShellAvailable[shell] = strings.TrimSpace(string(version))
	}

	seen := make(map[string]struct{}, len(tools))
	for _, tool := range tools {
		if _, exists := seen[tool]; exists {
			continue
		}
		seen[tool] = struct{}{}

		path, err := ResolveExecutable(tool)
		if err != nil {
			continue
		}
		absolutePath, err := filepath.Abs(path)
		if err != nil {
			continue
		}
		diagnostics.ResolvedTools[tool] = absolutePath
	}

	return diagnostics
}

// RedactKnownValues replaces every occurrence of known non-empty environment
// values at least eight characters long in text with a fixed placeholder.
func RedactKnownValues(text string, environment map[string]string) string {
	values := make([]string, 0, len(environment))
	seen := make(map[string]struct{}, len(environment))
	for _, value := range environment {
		if utf8.RuneCountInString(value) < 8 {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		values = append(values, value)
	}
	sort.Slice(values, func(i, j int) bool {
		return len(values[i]) > len(values[j])
	})
	for _, value := range values {
		text = strings.ReplaceAll(text, value, redactedValue)
	}
	return text
}

func environmentVariableNames() []string {
	names := make([]string, 0)
	for _, entry := range os.Environ() {
		name, _, found := strings.Cut(entry, "=")
		if found {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}
