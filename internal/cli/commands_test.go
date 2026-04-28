package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/larsartmann/oxlint-auto-configure/pkg/config"
	"github.com/larsartmann/oxlint-auto-configure/pkg/profile"
	"github.com/larsartmann/oxlint-auto-configure/pkg/rule"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigureDryRunRecommended(t *testing.T) {
	t.Parallel()
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"configure", "--dry-run", "--root", t.TempDir()})

	err := cmd.Execute()
	require.NoError(t, err)
}

func TestConfigureDryRunMaximalTypesafe(t *testing.T) {
	t.Parallel()
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"configure", "--dry-run", "-p", "maximal-typesafe", "--root", t.TempDir()})

	err := cmd.Execute()
	require.NoError(t, err)
}

func TestConfigureInvalidProfile(t *testing.T) {
	t.Parallel()
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"configure", "-p", "invalid", "--root", t.TempDir()})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid profile")
}

func TestConfigureWritesFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".oxlintrc.json")

	cmd := NewRootCommand()
	cmd.SetArgs([]string{"configure", "-c", configPath, "--root", dir})

	err := cmd.Execute()
	require.NoError(t, err)

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	var cfg config.OxlintConfig
	err = json.Unmarshal(data, &cfg)
	require.NoError(t, err)
	assert.Equal(t, "error", cfg.Categories["correctness"])
}

func TestValidateConfig(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".oxlintrc.json")

	reg, err := rule.LoadRegistry()
	require.NoError(t, err)

	cat := profile.NewCategorizer(profile.ProfileRecommended, profile.PluginConfig{})
	gen := config.NewGenerator(cat, reg)
	cfg := gen.Generate()
	data, err := cfg.ToJSON()
	require.NoError(t, err)
	err = os.WriteFile(configPath, append(data, '\n'), 0o644)
	require.NoError(t, err)

	cmd := NewRootCommand()
	cmd.SetArgs([]string{"validate", "-c", configPath})

	err = cmd.Execute()
	require.NoError(t, err)
}

func TestValidateUnknownRules(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".oxlintrc.json")

	badConfig := `{"rules":{"nonexistent-rule-xyz":"error"}}`
	err := os.WriteFile(configPath, []byte(badConfig), 0o644)
	require.NoError(t, err)

	cmd := NewRootCommand()
	cmd.SetArgs([]string{"validate", "-c", configPath})

	err = cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown rules")
}

func TestReportTable(t *testing.T) {
	t.Parallel()
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"report", "-f", "table", "--root", t.TempDir()})

	err := cmd.Execute()
	require.NoError(t, err)
}

func TestAnalyzeCleanProject(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"analyze", "--root", dir})

	err := cmd.Execute()
	require.NoError(t, err)
}

func TestVersionFlag(t *testing.T) {
	t.Parallel()
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"--version"})

	err := cmd.Execute()
	require.NoError(t, err)
}

func TestReportSummary(t *testing.T) {
	t.Parallel()
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"report", "-f", "summary", "--root", t.TempDir()})

	err := cmd.Execute()
	require.NoError(t, err)
}

func TestReportJSON(t *testing.T) {
	t.Parallel()
	reg, err := rule.LoadRegistry()
	require.NoError(t, err)

	cat := profile.NewCategorizer(profile.ProfileRecommended, profile.PluginConfig{})
	decisions := cat.DecideAll(reg)

	// Verify we can generate the report data
	assert.Len(t, decisions, 716)

	// Verify JSON serialization
	entries := make([]map[string]string, 0, len(decisions))
	for _, d := range decisions {
		entries = append(entries, map[string]string{
			"rule":     d.Rule.FullName(),
			"severity": string(d.Severity),
		})
	}
	data, err := json.Marshal(entries)
	require.NoError(t, err)
	assert.Greater(t, len(data), 1000)
}
