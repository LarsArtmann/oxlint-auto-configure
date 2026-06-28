package rule

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadTestRegistry(t *testing.T) *Registry {
	t.Helper()
	reg, err := LoadRegistry()
	require.NoError(t, err)
	return reg
}

func TestLoadRegistry(t *testing.T) {
	t.Parallel()
	reg := loadTestRegistry(t)
	assert.NotEmpty(t, reg.All())
}

func TestRegistryTotal(t *testing.T) {
	t.Parallel()
	reg := loadTestRegistry(t)
	assert.Equal(t, 716, reg.Len())
}

func TestRegistryByName(t *testing.T) {
	t.Parallel()
	reg := loadTestRegistry(t)

	r, ok := reg.ByName("no-unused-vars")
	assert.True(t, ok)
	assert.Equal(t, "no-unused-vars", r.Name)
	assert.Equal(t, PluginESLint, r.Plugin)
	assert.Equal(t, CategoryCorrectness, r.Category)
	assert.True(t, r.Enabled)
}

func TestRegistryByNameWithTypeScript(t *testing.T) {
	t.Parallel()
	reg := loadTestRegistry(t)

	r, ok := reg.ByName("typescript/no-floating-promises")
	assert.True(t, ok)
	assert.Equal(t, PluginTypeScript, r.Plugin)
	assert.True(t, r.TypeAware)
}

func TestRegistryByNameNotFound(t *testing.T) {
	t.Parallel()
	reg := loadTestRegistry(t)

	_, ok := reg.ByName("nonexistent-rule")
	assert.False(t, ok)
}

func TestRegistryByCategory(t *testing.T) {
	t.Parallel()
	reg := loadTestRegistry(t)

	correctness := reg.ByCategory(CategoryCorrectness)
	assert.NotEmpty(t, correctness)
	for _, r := range correctness {
		assert.Equal(t, CategoryCorrectness, r.Category)
	}
}

func TestRegistryByPlugin(t *testing.T) {
	t.Parallel()
	reg := loadTestRegistry(t)

	ts := reg.ByPlugin(PluginTypeScript)
	assert.NotEmpty(t, ts)
	for _, r := range ts {
		assert.Equal(t, PluginTypeScript, r.Plugin)
	}
}

func TestRegistryEnabledByDefault(t *testing.T) {
	t.Parallel()
	reg := loadTestRegistry(t)

	enabled := reg.EnabledByDefault()
	assert.NotEmpty(t, enabled)
	for _, r := range enabled {
		assert.True(t, r.Enabled)
	}
}

func TestRegistryDisabledByDefault(t *testing.T) {
	t.Parallel()
	reg := loadTestRegistry(t)

	disabled := reg.DisabledByDefault()
	assert.NotEmpty(t, disabled)
	for _, r := range disabled {
		assert.False(t, r.Enabled)
	}
}

func TestRegistryFixable(t *testing.T) {
	t.Parallel()
	reg := loadTestRegistry(t)

	fixable := reg.Fixable()
	assert.NotEmpty(t, fixable)
	for _, r := range fixable {
		assert.True(t, r.IsFixable())
	}
}

func TestRegistryTypeAware(t *testing.T) {
	t.Parallel()
	reg := loadTestRegistry(t)

	ta := reg.TypeAwareRules()
	assert.NotEmpty(t, ta)
	for _, r := range ta {
		assert.True(t, r.TypeAware)
	}
}

func TestRuleFullName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		rule     Rule
		expected string
	}{
		{Rule{Name: "no-unused-vars", Plugin: PluginESLint}, "no-unused-vars"},
		{
			Rule{Name: "no-floating-promises", Plugin: PluginTypeScript},
			"typescript/no-floating-promises",
		},
		{Rule{Name: "alt-text", Plugin: PluginJSXA11y}, "jsx_a11y/alt-text"},
		{Rule{Name: "google-font-display", Plugin: PluginNextJS}, "nextjs/google-font-display"},
		{Rule{Name: "valid-describe-callback", Plugin: PluginJest}, "jest/valid-describe-callback"},
		{
			Rule{Name: "bad-array-method-on-arguments", Plugin: PluginOXC},
			"oxc/bad-array-method-on-arguments",
		},
		{
			Rule{Name: "no-await-in-promise-methods", Plugin: PluginUnicorn},
			"unicorn/no-await-in-promise-methods",
		},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.rule.FullName())
	}
}

func TestCategoryIsValid(t *testing.T) {
	t.Parallel()
	assert.True(t, CategoryCorrectness.IsValid())
	assert.True(t, CategorySuspicious.IsValid())
	assert.True(t, CategoryNursery.IsValid())
	assert.False(t, Category("invalid").IsValid())
}

func TestPluginIsValid(t *testing.T) {
	t.Parallel()
	assert.True(t, PluginESLint.IsValid())
	assert.True(t, PluginTypeScript.IsValid())
	assert.True(t, PluginVue.IsValid())
	assert.False(t, Plugin("invalid").IsValid())
}

func TestMapFix(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected FixCapability
	}{
		{"none", FixNone},
		{"pending", FixNone},
		{"fixable_fix", FixSafe},
		{"fixable_safe_fix_or_suggestion", FixSafe},
		{"conditional_safe_fix_or_suggestion", FixSafe},
		{"fixable_suggestion", FixSuggestion},
		{"conditional_suggestion", FixSuggestion},
		{"fixable_dangerous_fix", FixDangerous},
		{"fixable_dangerous_suggestion", FixDangerous},
		{"unknown", FixNone},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, mapFix(tt.input))
	}
}

func TestAllCategories(t *testing.T) {
	t.Parallel()
	cats := AllCategories()
	assert.Len(t, cats, 7)
}

func TestAllPlugins(t *testing.T) {
	t.Parallel()
	plugins := AllPlugins()
	assert.Len(t, plugins, 15)
}

func TestEmbeddedVersion(t *testing.T) {
	t.Parallel()
	ver := EmbeddedVersion()
	assert.NotEmpty(t, ver)
	assert.Equal(t, "1.59.0", ver)
}

func TestPluginCLIFlag(t *testing.T) {
	t.Parallel()
	tests := []struct {
		plugin   Plugin
		expected string
	}{
		{PluginESLint, ""},
		{PluginUnicorn, ""},
		{PluginTypeScript, ""},
		{PluginOXC, ""},
		{PluginReact, "--react-plugin"},
		{PluginVue, "--vue-plugin"},
		{PluginJest, "--jest-plugin"},
		{PluginVitest, "--vitest-plugin"},
		{PluginNextJS, "--nextjs-plugin"},
		{PluginJSXA11y, "--jsx-a11y-plugin"},
		{PluginReactPerf, "--react-perf-plugin"},
		{PluginImport, "--import-plugin"},
		{PluginPromise, "--promise-plugin"},
		{PluginNode, "--node-plugin"},
		{PluginJSDoc, "--jsdoc-plugin"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.plugin.CLIFlag())
	}
}

func TestPluginNeedsFlag(t *testing.T) {
	t.Parallel()
	assert.False(t, PluginESLint.NeedsFlag())
	assert.False(t, PluginTypeScript.NeedsFlag())
	assert.False(t, PluginUnicorn.NeedsFlag())
	assert.False(t, PluginOXC.NeedsFlag())
	assert.True(t, PluginReact.NeedsFlag())
	assert.True(t, PluginVue.NeedsFlag())
	assert.True(t, PluginJest.NeedsFlag())
}

func TestRegistryFilter(t *testing.T) {
	t.Parallel()
	reg := loadTestRegistry(t)

	fixed := reg.Filter(func(r Rule) bool { return r.Fix == FixSafe })
	assert.NotEmpty(t, fixed)
	for _, r := range fixed {
		assert.True(t, r.IsSafeFixable())
	}
}

func TestSeverityDecisionString(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "error", string(SeverityError))
	assert.Equal(t, "warn", string(SeverityWarn))
	assert.Equal(t, "off", string(SeverityOff))
}

func TestCategoryString(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "correctness", CategoryCorrectness.String())
	assert.Equal(t, "suspicious", CategorySuspicious.String())
	assert.Equal(t, "nursery", CategoryNursery.String())
}

func TestPluginString(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "eslint", PluginESLint.String())
	assert.Equal(t, "typescript", PluginTypeScript.String())
	assert.Equal(t, "react", PluginReact.String())
}

func TestSeverityDecisionStringMethod(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "error", SeverityError.String())
	assert.Equal(t, "warn", SeverityWarn.String())
	assert.Equal(t, "off", SeverityOff.String())
}

func TestAlwaysOnPluginsConsistentWithAllPlugins(t *testing.T) {
	t.Parallel()
	for _, p := range AllPlugins() {
		if alwaysOnPlugins[p] {
			assert.False(t, p.NeedsFlag(), "always-on plugin %q should not need flag", p)
		} else {
			assert.True(t, p.NeedsFlag(), "non-always-on plugin %q should need flag", p)
		}
	}
}

func TestCLIFlagMapConsistentWithAllPlugins(t *testing.T) {
	t.Parallel()
	for _, p := range AllPlugins() {
		if p.NeedsFlag() {
			assert.NotEmpty(t, p.CLIFlag(), "plugin %q needs flag but CLIFlag() is empty", p)
		}
	}
}

func TestAlwaysOnPluginsSubsetOfAllPlugins(t *testing.T) {
	t.Parallel()
	allPluginsSet := make(map[Plugin]bool, len(AllPlugins()))
	for _, p := range AllPlugins() {
		allPluginsSet[p] = true
	}
	for p := range alwaysOnPlugins {
		assert.True(t, allPluginsSet[p], "alwaysOnPlugins contains %q not in AllPlugins()", p)
	}
	for p := range cliFlagMap {
		assert.True(t, allPluginsSet[p], "cliFlagMap contains %q not in AllPlugins()", p)
	}
}

func TestSeverityDecisionIsValid(t *testing.T) {
	t.Parallel()
	assert.True(t, SeverityError.IsValid())
	assert.True(t, SeverityWarn.IsValid())
	assert.True(t, SeverityOff.IsValid())
	assert.False(t, SeverityDecision("invalid").IsValid())
}
