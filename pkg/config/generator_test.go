package config

import (
	"encoding/json"
	"testing"

	"github.com/larsartmann/oxlint-auto-configure/pkg/detect"
	"github.com/larsartmann/oxlint-auto-configure/pkg/profile"
	"github.com/larsartmann/oxlint-auto-configure/pkg/rule"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratorRecommended(t *testing.T) {
	t.Parallel()
	reg, err := rule.LoadRegistry()
	require.NoError(t, err)

	cat := profile.NewCategorizer(profile.ProfileRecommended, profile.PluginConfig{})
	gen := NewGenerator(cat, reg, nil)
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
	t.Parallel()
	reg, err := rule.LoadRegistry()
	require.NoError(t, err)

	cat := profile.NewCategorizer(profile.ProfileMaximalTypesafe, profile.PluginConfig{})
	gen := NewGenerator(cat, reg, nil)
	cfg := gen.GenerateMaximal()

	assert.Equal(t, "error", cfg.Categories["correctness"])
	assert.Equal(t, "error", cfg.Categories["suspicious"])
	assert.Equal(t, "error", cfg.Categories["style"])
	assert.Equal(t, "error", cfg.Categories["perf"])
	assert.Equal(t, "error", cfg.Categories["pedantic"])
	assert.Equal(t, "error", cfg.Categories["restriction"])
	assert.Equal(t, "warn", cfg.Categories["nursery"])
}

func TestGeneratorWithReactProject(t *testing.T) {
	t.Parallel()
	reg, err := rule.LoadRegistry()
	require.NoError(t, err)

	pc := profile.PluginConfig{
		rule.PluginReact:     true,
		rule.PluginJSXA11y:   true,
		rule.PluginReactPerf: true,
	}
	cat := profile.NewCategorizer(profile.ProfileRecommended, pc)
	gen := NewGenerator(cat, reg, []detect.ProjectType{detect.ProjectTypeReact})
	cfg := gen.Generate()

	assert.Contains(t, cfg.Plugins, "react")
	assert.Contains(t, cfg.Plugins, "jsx_a11y")
	assert.True(t, cfg.Env["builtin"])
	assert.False(t, cfg.Env["node"], "React without Node should not add node env")
}

func TestGeneratorWithAllPlugins(t *testing.T) {
	t.Parallel()
	reg, err := rule.LoadRegistry()
	require.NoError(t, err)

	pc := profile.PluginConfig{
		rule.PluginReact: true, rule.PluginNextJS: true, rule.PluginVue: true,
		rule.PluginJest: true, rule.PluginVitest: true, rule.PluginJSDoc: true,
		rule.PluginJSXA11y: true, rule.PluginNode: true, rule.PluginImport: true,
		rule.PluginPromise: true, rule.PluginReactPerf: true,
	}
	cat := profile.NewCategorizer(profile.ProfileRecommended, pc)
	gen := NewGenerator(cat, reg, []detect.ProjectType{detect.ProjectTypeNode})
	cfg := gen.Generate()

	assert.Contains(t, cfg.Plugins, "react")
	assert.Contains(t, cfg.Plugins, "vue")
	assert.Contains(t, cfg.Plugins, "jest")
	assert.Contains(t, cfg.Plugins, "vitest")
	assert.True(t, cfg.Env["node"], "Node project type should add node env")
}

func TestConfigToJSON(t *testing.T) {
	t.Parallel()
	cfg := &OxlintConfig{
		Plugins:    []string{"typescript", "unicorn"},
		Categories: map[string]string{"correctness": "error"},
		Rules:      map[string]string{"no-unused-vars": "error"},
		Env:        map[string]bool{"builtin": true},
	}

	data, err := cfg.ToJSON()
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	var parsed map[string]any
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)
}

func TestConfigRoundTrip(t *testing.T) {
	t.Parallel()
	reg, err := rule.LoadRegistry()
	require.NoError(t, err)

	cat := profile.NewCategorizer(profile.ProfileRecommended, profile.PluginConfig{})
	gen := NewGenerator(cat, reg, nil)
	cfg := gen.Generate()

	data, err := cfg.ToJSON()
	require.NoError(t, err)

	parsed, err := FromJSON(data)
	require.NoError(t, err)

	assert.Equal(t, cfg.Categories, parsed.Categories)
}

func TestFromJSON(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	reg, err := rule.LoadRegistry()
	require.NoError(t, err)

	cat := profile.NewCategorizer(profile.ProfileMinimal, profile.PluginConfig{})
	gen := NewGenerator(cat, reg, nil)
	cfg := gen.Generate()

	assert.Equal(t, "error", cfg.Categories["correctness"])
	assert.NotContains(t, cfg.Categories, "style", "minimal should omit non-correctness categories")
	assert.NotContains(
		t,
		cfg.Categories,
		"suspicious",
		"minimal should omit non-correctness categories",
	)
	assert.NotContains(
		t,
		cfg.Categories,
		"nursery",
		"minimal should omit non-correctness categories",
	)
	assert.Empty(t, cfg.Rules, "minimal should have no per-rule overrides")
}

func TestRecommendedProfileNoRedundantOverrides(t *testing.T) {
	t.Parallel()
	reg, err := rule.LoadRegistry()
	require.NoError(t, err)

	cat := profile.NewCategorizer(profile.ProfileRecommended, profile.PluginConfig{})
	gen := NewGenerator(cat, reg, nil)
	cfg := gen.Generate()

	for ruleName, ruleSev := range cfg.Rules {
		r, ok := reg.ByName(ruleName)
		if !ok {
			continue
		}
		catSev, hasCat := cfg.Categories[string(r.Category)]
		if hasCat {
			assert.NotEqual(t, catSev, ruleSev,
				"rule %q should not duplicate its category severity", ruleName)
		}
	}
}

func TestFromJSONInvalid(t *testing.T) {
	t.Parallel()
	_, err := FromJSON([]byte("not json"))
	require.Error(t, err)
}

func TestFromJSONEmptyObject(t *testing.T) {
	t.Parallel()
	cfg, err := FromJSON([]byte("{}"))
	require.NoError(t, err)
	assert.Empty(t, cfg.Categories)
	assert.Empty(t, cfg.Rules)
}
