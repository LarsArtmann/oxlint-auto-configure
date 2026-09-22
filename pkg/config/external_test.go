package config

import (
	"testing"

	"github.com/larsartmann/oxlint-auto-configure/internal/testregistry"
	"github.com/larsartmann/oxlint-auto-configure/pkg/profile"
	"github.com/larsartmann/oxlint-auto-configure/pkg/rule"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// shadcnPlugin is the known external plugin used across these tests.
var shadcnPlugin = rule.ExternalPlugin{Package: "@shadcn/lint", Prefix: "shadcn"}

func TestFromJSONWithJsPluginsAndRuleOptions(t *testing.T) {
	t.Parallel()

	// The exact shape an @shadcn/lint setup produces: a jsPlugins
	// registration plus array-form rule values with options. This used to
	// fail unmarshaling (array into string) and made configure treat the
	// whole config as malformed.
	data := []byte(`{
		"jsPlugins": ["@shadcn/lint"],
		"rules": {
			"no-console": "error",
			"shadcn/no-restyle": ["error", {"allow": ["layout"]}],
			"shadcn/no-arbitrary-values": "error"
		},
		"settings": {
			"shadcn": {"ui": "@/components/ui"}
		}
	}`)

	cfg, err := FromJSON(data)
	require.NoError(t, err)

	assert.Equal(t, []string{"@shadcn/lint"}, cfg.JsPlugins)
	assert.Equal(t, "error", cfg.Rules["no-console"])
	assert.Equal(t, "error", cfg.Rules["shadcn/no-arbitrary-values"])

	options, ok := cfg.Rules["shadcn/no-restyle"].([]any)
	require.True(t, ok, "array-form rule value should round-trip as []any")
	require.Len(t, options, 2)
	assert.Equal(t, "error", options[0])

	shadcnSettings, ok := cfg.Settings["shadcn"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "@/components/ui", shadcnSettings["ui"])
}

func TestToJSONOmitsEmptyJsPlugins(t *testing.T) {
	t.Parallel()

	cfg := &OxlintConfig{Rules: map[string]any{"no-console": SeverityError}}

	data, err := cfg.ToJSON()
	require.NoError(t, err)
	assert.NotContains(t, string(data), "jsPlugins",
		"projects without external plugins must not get a jsPlugins key")
}

func TestGenerateRegistersDetectedExternalPlugins(t *testing.T) {
	t.Parallel()
	reg := testregistry.Load(t)

	cat := profile.NewCategorizer(profile.ProfileRecommended, profile.PluginConfig{})
	gen := NewGenerator(cat, reg, nil, []rule.ExternalPlugin{shadcnPlugin})
	cfg := gen.Generate()

	assert.Equal(t, []string{"@shadcn/lint"}, cfg.JsPlugins)
	assert.False(t, HasExternalRules(cfg),
		"detection must register the plugin without enabling any of its rules")

	roundTripped, err := FromJSON(mustJSON(t, cfg))
	require.NoError(t, err)
	assert.Equal(t, []string{"@shadcn/lint"}, roundTripped.JsPlugins)
}

func TestGenerateWithoutExternalPlugins(t *testing.T) {
	t.Parallel()
	reg := testregistry.Load(t)

	cat := profile.NewCategorizer(profile.ProfileRecommended, profile.PluginConfig{})
	cfg := NewGenerator(cat, reg, nil, nil).Generate()

	assert.Empty(t, cfg.JsPlugins)
}

func TestGenerateMaximalRegistersExternalPlugins(t *testing.T) {
	t.Parallel()
	reg := testregistry.Load(t)

	cat := profile.NewCategorizer(profile.ProfileMaximalTypesafe, profile.PluginConfig{})
	cfg := NewGenerator(cat, reg, nil, []rule.ExternalPlugin{shadcnPlugin}).GenerateMaximal()

	assert.Equal(t, []string{"@shadcn/lint"}, cfg.JsPlugins)
}

func TestPreserveExternalFullSetup(t *testing.T) {
	t.Parallel()

	existing := &OxlintConfig{
		JsPlugins: []string{"@shadcn/lint"},
		Rules: map[string]any{
			"shadcn/no-restyle":               []any{"error", map[string]any{"allow": []any{"layout"}}},
			"shadcn/no-arbitrary-values":      "error",
			"typescript/no-floating-promises": "error",
			"no-console":                      "off",
		},
		Settings: map[string]any{
			"shadcn": map[string]any{"ui": "@/components/ui"},
			"react":  map[string]any{"version": "detect"},
		},
	}
	generated := &OxlintConfig{
		Plugins: []string{"typescript"},
		Rules:   map[string]any{"no-unused-vars": SeverityError},
		Settings: map[string]any{
			"vitest": map[string]any{"typecheck": false},
		},
	}

	merged := PreserveExternal(existing, generated)

	assert.Equal(t, []string{"@shadcn/lint"}, merged.JsPlugins)

	// External rules preserved verbatim, including options...
	preserved, ok := merged.Rules["shadcn/no-restyle"].([]any)
	require.True(t, ok)
	require.Len(t, preserved, 2)
	assert.Equal(t, map[string]any{"allow": []any{"layout"}}, preserved[1])
	assert.Equal(t, "error", merged.Rules["shadcn/no-arbitrary-values"])

	// ...while built-in rules follow the generated config, not the old one.
	assert.Equal(t, SeverityError, merged.Rules["no-unused-vars"])
	assert.NotContains(t, merged.Rules, "typescript/no-floating-promises")
	assert.NotContains(t, merged.Rules, "no-console")

	// External settings preserved; other settings stay generated.
	assert.Equal(t, map[string]any{"ui": "@/components/ui"}, merged.Settings["shadcn"])
	assert.NotContains(t, merged.Settings, "react")
	assert.Contains(t, merged.Settings, "vitest")
}

func TestPreserveExternalUnionsJsPlugins(t *testing.T) {
	t.Parallel()

	existing := &OxlintConfig{JsPlugins: []string{"@other/pkg"}}
	generated := &OxlintConfig{JsPlugins: []string{"@shadcn/lint"}}

	merged := PreserveExternal(existing, generated)

	assert.Equal(t, []string{"@other/pkg", "@shadcn/lint"}, merged.JsPlugins,
		"existing registrations are kept even when the package left the deps")
}

func TestPreserveExternalNilExisting(t *testing.T) {
	t.Parallel()

	generated := &OxlintConfig{Rules: map[string]any{"no-console": SeverityError}}

	assert.Same(t, generated, PreserveExternal(nil, generated))
}

func TestPreserveExternalNilGeneratedRules(t *testing.T) {
	t.Parallel()

	existing := &OxlintConfig{Rules: map[string]any{"shadcn/no-raw-colors": "error"}}
	generated := &OxlintConfig{}

	merged := PreserveExternal(existing, generated)

	assert.Equal(t, "error", merged.Rules["shadcn/no-raw-colors"])
}

func TestFromJSONWithOverrides(t *testing.T) {
	t.Parallel()

	// The shape @shadcn/lint's setup produces: per-glob overrides disabling
	// rules inside component directories. Regeneration used to drop this
	// silently.
	data := []byte(`{
		"rules": {"no-console": "error"},
		"overrides": [
			{
				"files": ["src/components/ui/**"],
				"rules": {"shadcn/no-restyle": "off"}
			}
		]
	}`)

	cfg, err := FromJSON(data)
	require.NoError(t, err)

	require.Len(t, cfg.Overrides, 1)
	assert.Equal(t, []any{"src/components/ui/**"}, cfg.Overrides[0]["files"])

	rules, ok := cfg.Overrides[0]["rules"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "off", rules["shadcn/no-restyle"])

	roundTripped, err := FromJSON(mustJSON(t, cfg))
	require.NoError(t, err)
	assert.Equal(t, cfg.Overrides, roundTripped.Overrides)
}

func TestToJSONOmitsEmptyOverrides(t *testing.T) {
	t.Parallel()

	cfg := &OxlintConfig{Rules: map[string]any{"no-console": SeverityError}}

	data, err := cfg.ToJSON()
	require.NoError(t, err)
	assert.NotContains(t, string(data), "overrides",
		"projects without overrides must not get an overrides key")
}

func TestPreserveExternalCopiesOverridesVerbatim(t *testing.T) {
	t.Parallel()

	existing := &OxlintConfig{
		Overrides: []map[string]any{
			{
				"files": []any{"src/components/ui/**"},
				"rules": map[string]any{"shadcn/no-restyle": "off"},
			},
			{
				"files": []any{"*.config.js"},
				"rules": map[string]any{"no-console": "off"},
			},
		},
	}
	generated := &OxlintConfig{Rules: map[string]any{"no-console": SeverityError}}

	merged := PreserveExternal(existing, generated)

	assert.Equal(t, existing.Overrides, merged.Overrides,
		"overrides are pure existing policy and must survive regeneration verbatim")
}

func TestPreserveExternalDedupsOverrides(t *testing.T) {
	t.Parallel()

	duplicate := map[string]any{
		"files": []any{"src/components/ui/**"},
		"rules": map[string]any{"shadcn/no-restyle": "off"},
	}
	distinct := map[string]any{
		"files": []any{"*.config.js"},
		"rules": map[string]any{"no-console": "off"},
	}
	// Same content as duplicate but different key order: must still dedup.
	reordered := map[string]any{
		"rules": map[string]any{"shadcn/no-restyle": "off"},
		"files": []any{"src/components/ui/**"},
	}

	existing := &OxlintConfig{Overrides: []map[string]any{duplicate, distinct, duplicate}}
	generated := &OxlintConfig{Overrides: []map[string]any{reordered}}

	merged := PreserveExternal(existing, generated)

	assert.Equal(t, []map[string]any{reordered, distinct}, merged.Overrides,
		"exact duplicates collapse (key order irrelevant), distinct blocks survive in order")
}


func TestHasExternalRules(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		rules map[string]any
		want  bool
	}{
		{"empty", nil, false},
		{"builtin only", map[string]any{"no-console": "error"}, false},
		{
			"external present",
			map[string]any{"no-console": "error", "shadcn/no-restyle": "error"},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, HasExternalRules(&OxlintConfig{Rules: tt.rules}))
		})
	}
}

func mustJSON(t *testing.T, cfg *OxlintConfig) []byte {
	t.Helper()

	data, err := cfg.ToJSON()
	require.NoError(t, err)

	return data
}
