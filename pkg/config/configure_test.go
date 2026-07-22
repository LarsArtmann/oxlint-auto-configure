package config

import (
	"testing"

	"github.com/larsartmann/oxlint-auto-configure/pkg/detect"
	"github.com/larsartmann/oxlint-auto-configure/pkg/profile"
	"github.com/larsartmann/oxlint-auto-configure/pkg/rule"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadTestRegistry(t *testing.T) *rule.Registry {
	t.Helper()

	reg, err := rule.LoadRegistry()
	require.NoError(t, err)

	return reg
}

func TestGenerateProjectConfigRecommended(t *testing.T) {
	t.Parallel()
	reg := loadTestRegistry(t)

	cfg, err := GenerateProjectConfig(profile.ProfileRecommended, reg, nil, nil)
	require.NoError(t, err)
	assert.Contains(t, cfg.Plugins, "typescript")
	assert.Equal(t, "error", cfg.Categories["correctness"])
}

func TestGenerateProjectConfigMaximal(t *testing.T) {
	t.Parallel()
	reg := loadTestRegistry(t)

	cfg, err := GenerateProjectConfig(profile.ProfileMaximalTypesafe, reg, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "error", cfg.Categories["correctness"])
	assert.Equal(t, "warn", cfg.Categories["nursery"])
}

func TestGenerateProjectConfigInvalidProfile(t *testing.T) {
	t.Parallel()
	reg := loadTestRegistry(t)

	_, err := GenerateProjectConfig(profile.Profile("invalid"), reg, nil, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidProfile)
}

func TestGenerateProjectConfigWithNodeProject(t *testing.T) {
	t.Parallel()
	reg := loadTestRegistry(t)

	cfg, err := GenerateProjectConfig(
		profile.ProfileRecommended, reg, nil,
		[]detect.ProjectType{detect.ProjectTypeNode},
	)
	require.NoError(t, err)
	assert.True(t, cfg.Env["node"])
}

func TestGenerateProjectConfigWithReactPlugins(t *testing.T) {
	t.Parallel()
	reg := loadTestRegistry(t)

	pc := profile.PluginConfig{
		rule.PluginReact:   true,
		rule.PluginJSXA11y: true,
	}
	cfg, err := GenerateProjectConfig(profile.ProfileRecommended, reg, pc, nil)
	require.NoError(t, err)
	assert.Contains(t, cfg.Plugins, "react")
	assert.Contains(t, cfg.Plugins, "jsx_a11y")
}
