package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/larsartmann/oxlint-auto-configure/pkg/config"
	"github.com/larsartmann/oxlint-auto-configure/pkg/profile"
	"github.com/larsartmann/oxlint-auto-configure/pkg/rule"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWriteConfigLeavesNoTempFiles verifies that writeConfig uses atomic
// writes (temp + rename) and leaves no .tmp files behind.
func TestWriteConfigLeavesNoTempFiles(t *testing.T) {
	t.Parallel()

	reg, err := rule.LoadRegistry()
	require.NoError(t, err)

	cfg, err := config.GenerateProjectProfile(profile.ProfileRecommended, reg, nil, nil)
	require.NoError(t, err)

	dir := t.TempDir()
	targetPath := filepath.Join(dir, ".oxlintrc.json")

	err = writeConfig(cfg, targetPath)
	require.NoError(t, err)

	// Verify the config file exists and is valid JSON
	data, err := os.ReadFile(targetPath)
	require.NoError(t, err, "config must exist after writeConfig")

	parsed, err := config.FromJSON(data)
	require.NoError(t, err, "config must be valid JSON")
	assert.NotEmpty(t, parsed.Categories)

	// Verify NO .tmp files left behind
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	for _, entry := range entries {
		assert.False(t, strings.HasSuffix(entry.Name(), ".tmp"),
			"no .tmp files should remain after atomic write, found: %s", entry.Name())
	}
}

// TestWriteConfigOverwriteIdempotent verifies that calling writeConfig twice
// to the same path produces valid output both times with no leftover temp files.
func TestWriteConfigOverwriteIdempotent(t *testing.T) {
	t.Parallel()

	reg, err := rule.LoadRegistry()
	require.NoError(t, err)

	cfg, err := config.GenerateProjectProfile(profile.ProfileRecommended, reg, nil, nil)
	require.NoError(t, err)

	dir := t.TempDir()
	targetPath := filepath.Join(dir, ".oxlintrc.json")

	// Write once
	require.NoError(t, writeConfig(cfg, targetPath))

	firstData, err := os.ReadFile(targetPath)
	require.NoError(t, err)

	// Write again (overwrite)
	require.NoError(t, writeConfig(cfg, targetPath))

	secondData, err := os.ReadFile(targetPath)
	require.NoError(t, err)

	// Both writes should produce identical output
	assert.Equal(t, firstData, secondData, "overwriting with same config should produce identical output")

	// Verify no temp files remain
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	for _, entry := range entries {
		assert.False(t, strings.HasSuffix(entry.Name(), ".tmp"),
			"no .tmp files should remain after overwrite, found: %s", entry.Name())
	}
}
