package oxlint

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunFixRemovesDebugger(t *testing.T) {
	t.Parallel()
	if os.Getenv("OXLINT_E2E") == "" {
		t.Skip("Set OXLINT_E2E=1 to run e2e test with real oxlint")
	}

	dir := t.TempDir()
	jsPath := filepath.Join(dir, "test.js")
	err := os.WriteFile(jsPath, []byte("debugger;\n"), 0o644)
	require.NoError(t, err)

	result, err := RunFix(context.Background(), dir, "")
	require.NoError(t, err)
	assert.NotNil(t, result)

	// oxlint --fix should have removed the debugger statement
	content, err := os.ReadFile(jsPath)
	require.NoError(t, err)
	assert.NotContains(t, string(content), "debugger")
}

func TestRunFixNoIssues(t *testing.T) {
	t.Parallel()
	if os.Getenv("OXLINT_E2E") == "" {
		t.Skip("Set OXLINT_E2E=1 to run e2e test with real oxlint")
	}

	dir := t.TempDir()
	jsPath := filepath.Join(dir, "clean.js")
	err := os.WriteFile(jsPath, []byte("const x = 1;\n"), 0o644)
	require.NoError(t, err)

	result, err := RunFix(context.Background(), dir, "")
	require.NoError(t, err)
	assert.NotNil(t, result)

	// File should be unchanged
	content, err := os.ReadFile(jsPath)
	require.NoError(t, err)
	assert.Equal(t, "const x = 1;\n", string(content))
}

func TestRunFixWithConfig(t *testing.T) {
	t.Parallel()
	if os.Getenv("OXLINT_E2E") == "" {
		t.Skip("Set OXLINT_E2E=1 to run e2e test with real oxlint")
	}

	dir := t.TempDir()
	configPath := filepath.Join(dir, ".oxlintrc.json")
	err := os.WriteFile(configPath, []byte(`{"rules":{"no-debugger":"error"}}`), 0o644)
	require.NoError(t, err)

	jsPath := filepath.Join(dir, "test.js")
	err = os.WriteFile(jsPath, []byte("debugger;\n"), 0o644)
	require.NoError(t, err)

	result, err := RunFix(context.Background(), dir, configPath)
	require.NoError(t, err)
	assert.NotNil(t, result)

	content, err := os.ReadFile(jsPath)
	require.NoError(t, err)
	assert.NotContains(t, string(content), "debugger")
}
