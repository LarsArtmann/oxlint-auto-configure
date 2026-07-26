package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestMainVersion builds the binary and verifies --version works.
// This provides coverage for the entry point (main.go) which currently has 0% coverage.
func TestMainVersion(t *testing.T) {
	t.Parallel()

	binary := buildBinary(t)

	output, err := exec.Command(binary, "--version").Output()
	if err != nil {
		t.Fatalf("binary --version failed: %v", err)
	}

	if len(output) == 0 {
		t.Fatal("expected non-empty version output")
	}
}

// TestMainHelp verifies the binary responds to --help without error.
func TestMainHelp(t *testing.T) {
	t.Parallel()

	binary := buildBinary(t)

	err := exec.Command(binary, "--help").Run()
	if err != nil {
		// cobra returns nil for --help, but some environments may differ
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() != 0 {
			t.Fatalf("binary --help failed with exit code %d", exitErr.ExitCode())
		}
	}
}

func buildBinary(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	binary := filepath.Join(dir, "oxlint-auto-configure-test")

	cmd := exec.Command("go", "build", "-o", binary, ".")
	cmd.Env = append(os.Environ(), "GOWORK=off", "GOEXPERIMENT=jsonv2")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, output)
	}

	return binary
}
