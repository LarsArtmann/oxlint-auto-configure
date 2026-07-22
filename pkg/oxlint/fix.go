package oxlint

import (
	"context"
	"os/exec"
	"strings"
)

// FixResult holds the output from running oxlint --fix.
type FixResult struct {
	Output string
}

// RunFix executes oxlint --fix in the given root directory.
// Returns the combined stdout/stderr output. Non-zero exit codes
// from oxlint are treated as findings, not failures.
func RunFix(ctx context.Context, rootDir, configPath string) (*FixResult, error) {
	args := []string{"--fix"}
	if configPath != "" {
		args = append(args, "-c", configPath)
	}

	args = append(args, ".")

	cmd := exec.CommandContext(ctx, "oxlint", args...)
	cmd.Dir = rootDir

	output, err := cmd.CombinedOutput()
	result := &FixResult{Output: strings.TrimSpace(string(output))}

	if err := checkExitError(err, "oxlint --fix"); err != nil {
		return nil, err
	}

	return result, nil
}
