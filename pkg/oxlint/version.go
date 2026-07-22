package oxlint

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

var versionRegex = regexp.MustCompile(`(\d+\.\d+\.\d+)`)

// CheckVersion runs `oxlint --version` and returns the version string.
func CheckVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "oxlint", "--version")

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrNotFound, err)
	}

	version := strings.TrimSpace(string(output))

	match := versionRegex.FindString(version)
	if match == "" {
		return "", fmt.Errorf("unexpected oxlint version output: %s", version)
	}

	return match, nil
}

// CheckBinary verifies oxlint is available in PATH.
func CheckBinary(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "oxlint", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w in PATH: %w", ErrNotFound, err)
	}

	return nil
}
