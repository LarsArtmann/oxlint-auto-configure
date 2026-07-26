package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/larsartmann/oxlint-auto-configure/pkg/config"
	"github.com/larsartmann/oxlint-auto-configure/pkg/profile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigureE2ERoundTrip generates a config via Configure(), reads it back,
// and verifies it parses correctly via config.FromJSON — proving the generated
// output is always valid, parseable JSON.
func TestConfigureE2ERoundTrip(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// Create a fake package.json so detection has something to read
	err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{}`), 0o644)
	require.NoError(t, err)

	opts := ConfigureOptions{
		Profile:    profile.ProfileRecommended,
		ConfigPath: filepath.Join(dir, ".oxlintrc.json"),
	}

	err = Configure(context.Background(), dir, opts)
	require.NoError(t, err)

	// Read the generated config
	data, err := os.ReadFile(opts.ConfigPath)
	require.NoError(t, err, "config file must exist after Configure")

	require.NotEmpty(t, data, "config file must not be empty")

	// Parse it back — this is the round-trip verification
	cfg, err := config.FromJSON(data)
	require.NoError(t, err, "generated config must be valid JSON that round-trips")

	assert.NotEmpty(t, cfg.Categories, "config must have categories")
	assert.Contains(t, cfg.Categories, "correctness")
	assert.Equal(t, "error", cfg.Categories["correctness"])
}

// TestConfigureE2EAllProfiles verifies that every profile produces valid,
// parseable JSON output.
func TestConfigureE2EAllProfiles(t *testing.T) {
	t.Parallel()

	for _, p := range profile.AllProfileNames() {
		t.Run(string(p), func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()

			err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{}`), 0o644)
			require.NoError(t, err)

			configPath := filepath.Join(dir, ".oxlintrc.json")
			opts := ConfigureOptions{
				Profile:    p,
				ConfigPath: configPath,
			}

			err = Configure(context.Background(), dir, opts)
			require.NoError(t, err)

			data, err := os.ReadFile(configPath)
			require.NoError(t, err)

			cfg, err := config.FromJSON(data)
			require.NoError(t, err, "profile %s must produce parseable JSON", p)

			assert.NotEmpty(t, cfg.Categories, "profile %s must have categories", p)
		})
	}
}
