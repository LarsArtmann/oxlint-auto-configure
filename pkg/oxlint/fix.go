package oxlint

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type FixResult struct {
	FilesFixed int
	Output     string
}

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

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if len(exitErr.Stderr) > 0 {
				return nil, fmt.Errorf("oxlint --fix: %s", string(exitErr.Stderr))
			}
		} else {
			return nil, fmt.Errorf("run oxlint --fix: %w", err)
		}
	}

	return result, nil
}
