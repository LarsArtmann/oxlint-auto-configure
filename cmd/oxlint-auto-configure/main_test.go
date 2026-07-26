package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestMainVersion(t *testing.T) {
	t.Parallel()

	binary := buildBinary(t)
	ctx := context.Background()

	output, err := exec.CommandContext(ctx, binary, "--version").Output()
	if err != nil {
		t.Fatalf("binary --version failed: %v", err)
	}

	if len(output) == 0 {
		t.Fatal("expected non-empty version output")
	}
}

func TestMainHelp(t *testing.T) {
	t.Parallel()

	binary := buildBinary(t)
	ctx := context.Background()

	if err := exec.CommandContext(ctx, binary, "--help").Run(); err != nil {
		t.Fatalf("binary --help failed: %v", err)
	}
}

func buildBinary(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	binary := filepath.Join(dir, "oxlint-auto-configure-test")

	cmd := exec.CommandContext(context.Background(), "go", "build", "-o", binary, ".")

	cmd.Env = append(os.Environ(), "GOWORK=off", "GOEXPERIMENT=jsonv2")

	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, output)
	}

	return binary
}
