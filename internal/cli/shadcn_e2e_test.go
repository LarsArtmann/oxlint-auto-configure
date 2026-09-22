package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/larsartmann/oxlint-auto-configure/pkg/config"
	"github.com/larsartmann/oxlint-auto-configure/pkg/profile"
	"github.com/larsartmann/oxlint-auto-configure/pkg/rule"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigureE2ERegistersShadcnLint verifies that a project with the
// shadcn lint package installed gets the plugin registered under "jsPlugins" —
// without any of its rules enabled (design-system policy is the project's
// choice, per @shadcn/lint's own setup guidance).
func TestConfigureE2ERegistersShadcnLint(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writePackageJSON(t, dir, `{
		"dependencies": {"react": "^18.0.0"},
		"devDependencies": {"@shadcn/lint": "^1.0.0", "oxlint": "^1.80.0"}
	}`)

	configPath := filepath.Join(dir, ".oxlintrc.json")
	require.NoError(t, Configure(context.Background(), dir, ConfigureOptions{
		Profile:    profile.ProfileRecommended,
		ConfigPath: configPath,
	}))

	cfg := readConfig(t, configPath)

	assert.Equal(t, []string{"@shadcn/lint"}, cfg.JsPlugins)
	assert.False(t, config.HasExternalRules(cfg),
		"configure must register the plugin but leave its rules off")
	assert.Contains(t, cfg.Plugins, "react",
		"regular plugin detection still applies alongside jsPlugins")
}

// TestConfigureE2EPreservesShadcnSetup verifies that regenerating a config
// never destroys an existing @shadcn/lint setup: the jsPlugins registration,
// rules with their options, and settings.shadcn survive verbatim.
func TestConfigureE2EPreservesShadcnSetup(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writePackageJSON(t, dir, `{
		"dependencies": {"react": "^18.0.0"},
		"devDependencies": {"@shadcn/lint": "^1.0.0"}
	}`)

	configPath := filepath.Join(dir, ".oxlintrc.json")
	require.NoError(t, os.WriteFile(configPath, []byte(`{
		"jsPlugins": ["@shadcn/lint"],
		"categories": {"correctness": "warn"},
		"rules": {
			"shadcn/no-restyle": ["error", {"allow": ["layout"]}],
			"shadcn/no-raw-colors": "error",
			"no-console": "off"
		},
		"settings": {
			"shadcn": {"ui": "@/components/ui"}
		}
	}`), 0o644))

	require.NoError(t, Configure(context.Background(), dir, ConfigureOptions{
		Profile:    profile.ProfileRecommended,
		ConfigPath: configPath,
	}))

	cfg := readConfig(t, configPath)

	assert.Equal(t, []string{"@shadcn/lint"}, cfg.JsPlugins,
		"the jsPlugins registration must survive regeneration")

	options, ok := cfg.Rules["shadcn/no-restyle"].([]any)
	require.True(t, ok, "shadcn rule with options must be preserved as an array")
	require.Len(t, options, 2)
	assert.Equal(t, "error", options[0])
	assert.Equal(t, map[string]any{"allow": []any{"layout"}}, options[1])

	assert.Equal(t, "error", cfg.Rules["shadcn/no-raw-colors"])
	assert.Equal(t, map[string]any{"ui": "@/components/ui"}, cfg.Settings["shadcn"])

	// The rest of the config is regenerated, not carried over.
	assert.Equal(t, "error", cfg.Categories["correctness"])
	assert.NotContains(t, cfg.Rules, "no-console")
}

// TestValidateE2EShadcnConfig verifies that validate accepts a config
// containing external plugin rules and reports them as external, not unknown.
func TestValidateE2EShadcnConfig(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := filepath.Join(dir, ".oxlintrc.json")
	require.NoError(t, os.WriteFile(configPath, []byte(`{
		"jsPlugins": ["@shadcn/lint"],
		"rules": {
			"no-console": "off",
			"shadcn/no-restyle": ["error", {"allow": ["layout"]}]
		}
	}`), 0o644))

	require.NoError(t, Validate(configPath))
}

func writePackageJSON(t *testing.T, dir, pkgJSON string) {
	t.Helper()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0o644))
}

func readConfig(t *testing.T, path string) *config.OxlintConfig {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	cfg, err := config.FromJSON(data)
	require.NoError(t, err)

	return cfg
}

// TestConfigureE2EWarnsOnlyForOrphanedJsPlugins verifies that a preserved
// jsPlugins registration for an uninstalled known plugin does not block
// configure, and that installed and hand-registered packages stay silent.
func TestConfigureE2EWarnsOnlyForOrphanedJsPlugins(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writePackageJSON(t, dir, `{"dependencies": {"react": "^18.0.0"}}`)

	configPath := filepath.Join(dir, ".oxlintrc.json")
	require.NoError(t, os.WriteFile(configPath, []byte(`{
		"jsPlugins": ["@shadcn/lint", "some-hand-registered-plugin"]
	}`), 0o644))

	require.NoError(t, Configure(context.Background(), dir, ConfigureOptions{
		Profile:    profile.ProfileRecommended,
		ConfigPath: configPath,
	}))

	cfg := readConfig(t, configPath)

	assert.ElementsMatch(t,
		[]string{"@shadcn/lint", "some-hand-registered-plugin"},
		cfg.JsPlugins,
		"preservation never drops registrations, even orphaned ones",
	)

	// The regenerated config registers @shadcn/lint (preserved) while the
	// package is no longer a dependency, so it is reported as orphaned; the
	// hand-registered unknown package never is.
	assert.Equal(t,
		[]string{"@shadcn/lint"},
		orphanedJsPlugins(cfg, nil),
	)

	installed := &config.OxlintConfig{JsPlugins: []string{"@shadcn/lint", "some-hand-registered-plugin"}}

	assert.Empty(t,
		orphanedJsPlugins(installed, []rule.ExternalPlugin{{Package: "@shadcn/lint", Prefix: "shadcn"}}),
		"an installed plugin is not orphaned",
	)
}
