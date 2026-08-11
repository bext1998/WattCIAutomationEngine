package env

import (
	"fmt"
	"os/exec"
)

// ResolveExecutable resolves an executable name using the current PATH.
func ResolveExecutable(name string) (string, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", fmt.Errorf("executable %q is not available in PATH: %w", name, err)
	}

	return path, nil
}
