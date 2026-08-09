package env

import "path/filepath"

// ResolveCwd resolves a step cwd relative to the repository root.
func ResolveCwd(repoRoot, stepCwd string) string {
	if stepCwd == "" {
		return repoRoot
	}
	if filepath.IsAbs(stepCwd) {
		return stepCwd
	}

	return filepath.Join(repoRoot, stepCwd)
}
