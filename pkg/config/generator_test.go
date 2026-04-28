package config

import (
	"encoding/json"
	"testing"

	"github.com/larsartmann/oxlint-auto-configure/pkg/profile"
	"github.com/larsartmann/oxlint-auto-configure/pkg/rule"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratorRecommended(t *testing.T) {
	reg, err := rule.LoadRegistry()
	require.NoError(t, err)

	cat := profile.NewCategorizer(profile.ProfileRecommended, profile.PluginConfig{})
	gen := NewGenerator(cat, reg)
	cfg := gen.Generate()

	assert.Contains(t, cfg.Plugins, "typescript")
	assert.Contains(t, cfg.Plugins, "unicorn")
	assert.Contains(t, cfg.Plugins, "oxc")

	assert.Equal(t, "error", cfg.Categories["correctness"])
	assert.Equal(t, "error", cfg.Categories["suspicious"])
	assert.Equal(t, "warn", cfg.Categories["style"])
	assert.Equal(t, "warn", cfg.Categories["perf"])
	assert.Equal(t, "warn", cfg.Categories["pedantic"])
	assert.Equal(t, "warn", cfg.Categories["restriction"])
	assert.Equal(t, "off", cfg.Categories["nursery"])
}

func TestGeneratorMaximalTypesafe(t *testing.T) {
	reg, err := rule.LoadRegistry()
	require.NoError(t, err)

	cat := profile.NewCategorizer(profile.ProfileMaximalTypesafe, profile.PluginConfig{})
	gen := NewGenerator(cat, reg)
	cfg := gen.GenerateAllError()

	assert.Equal(t, "error", cfg.Categories["correctness"])
	assert.Equal(t, "error", cfg.Categories["suspicious"])
	assert.Equal(t, "error", cfg.Categories["style"])
	assert.Equal(t, "error", cfg.Categories["perf"])
	assert.Equal(t, "error", cfg.Categories["pedantic"])
	assert.Equal(t, "error", cfg.Categories["restriction"])
	assert.Equal(t, "warn", cfg.Categories["nursery"])
}

func TestGeneratorWithReactProject(t *testing.T) {
	reg, err := rule.LoadRegistry()
	require.NoError(t, err)

	pc := profile.PluginConfig{React: true, JSXA11y: true, ReactPerf: true}
	cat := profile.NewCategorizer(profile.ProfileRecommended, pc)
	gen := NewGenerator(cat, reg)
	cfg := gen.Generate()

	assert.Contains(t, cfg.Plugins, "react")
	assert.Contains(t, cfg.Plugins, "jsx_a11y")
}

func TestGeneratorWithAllPlugins(t *testing.T) {
	reg, err := rule.LoadRegistry()
	require.NoError(t, err)

	pc := profile.PluginConfig{
		React: true, NextJS: true, Vue: true, Jest: true,
		Vitest: true, JSDoc: true, JSXA11y: true, Node: true,
		Import: true, Promise: true, ReactPerf: true,
	}
	cat := profile.NewCategorizer(profile.ProfileRecommended, pc)
	gen := NewGenerator(cat, reg)
	cfg := gen.Generate()

	assert.Contains(t, cfg.Plugins, "react")
	assert.Contains(t, cfg.Plugins, "vue")
	assert.Contains(t, cfg.Plugins, "jest")
	assert.Contains(t, cfg.Plugins, "vitest")
}

func TestConfigToJSON(t *testing.T) {
	cfg := &OxlintConfig{
		Plugins:  []string{"typescript", "unicorn"},
		Categories: map[string]string{"correctness": "error"},
		Rules:    map[string]string{"no-unused-vars": "error"},
		Env:      map[string]bool{"builtin": true},
	}

	data, err := cfg.ToJSON()
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	var parsed map[string]interface{}
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)
}

func TestConfigRoundTrip(t *testing.T) {
	reg, err := rule.LoadRegistry()
	require.NoError(t, err)

	cat := profile.NewCategorizer(profile.ProfileRecommended, profile.PluginConfig{})
	gen := NewGenerator(cat, reg)
	cfg := gen.Generate()

	data, err := cfg.ToJSON()
	require.NoError(t, err)

	parsed, err := FromJSON(data)
	require.NoError(t, err)

	assert.Equal(t, cfg.Categories, parsed.Categories)
}

func TestFromJSON(t *testing.T) {
	input := `{
		"plugins": ["typescript"],
		"categories": {"correctness": "error"},
		"rules": {"no-unused-vars": "error"},
		"env": {"builtin": true}
	}`

	cfg, err := FromJSON([]byte(input))
	require.NoError(t, err)
	assert.Contains(t, cfg.Plugins, "typescript")
	assert.Equal(t, "error", cfg.Categories["correctness"])
	assert.Equal(t, "error", cfg.Rules["no-unused-vars"])
}

func TestMinimalProfileConfig(t *testing.T) {
	reg, err := rule.LoadRegistry()
	require.NoError(t, err)

	cat := profile.NewCategorizer(profile.ProfileMinimal, profile.PluginConfig{})
	gen := NewGenerator(cat, reg)
	cfg := gen.Generate()

	assert.Equal(t, "error", cfg.Categories["correctness"])
}
