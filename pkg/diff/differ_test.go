package diff

import (
	"testing"

	"github.com/larsartmann/oxlint-auto-configure/pkg/config"
	"github.com/stretchr/testify/assert"
)

const (
	testSeverityError    = "error"
	testSeverityWarn     = "warn"
	testSeverityOff      = "off"
	testPluginTS         = "typescript"
	testPluginReact      = "react"
	testPluginVue        = "vue"
	testRuleNoUnusedVars = "no-unused-vars"
)

func TestDiffNoChanges(t *testing.T) {
	t.Parallel()

	before := &config.OxlintConfig{
		Rules: map[string]any{testRuleNoUnusedVars: testSeverityError},
	}
	after := &config.OxlintConfig{Rules: map[string]any{testRuleNoUnusedVars: testSeverityError}}

	d := NewDiffer(before, after)
	changes := d.Diff()
	assert.Empty(t, changes)
}

func TestDifferHasChanges(t *testing.T) {
	t.Parallel()

	identical := &config.OxlintConfig{
		Rules:     map[string]any{testRuleNoUnusedVars: testSeverityError},
		Plugins:   []string{testPluginTS},
		Overrides: []map[string]any{{"files": []string{"gen.ts"}, "rules": map[string]any{"no-explicit-any": "off"}}},
	}

	assert.False(t, NewDiffer(identical, identical).HasChanges(), "identical configs must not report changes")

	after := &config.OxlintConfig{
		Rules:     map[string]any{testRuleNoUnusedVars: testSeverityWarn},
		Plugins:   []string{testPluginTS},
		Overrides: []map[string]any{{"files": []string{"gen.ts"}, "rules": map[string]any{"no-explicit-any": "off"}}},
	}
	assert.True(t, NewDiffer(identical, after).HasChanges(), "a changed rule severity must report changes")
}

func TestDiffAdded(t *testing.T) {
	t.Parallel()

	before := &config.OxlintConfig{Rules: map[string]any{}}
	after := &config.OxlintConfig{Rules: map[string]any{testRuleNoUnusedVars: testSeverityError}}

	d := NewDiffer(before, after)
	changes := d.Diff()
	assert.Len(t, changes, 1)
	assert.Equal(t, KindAdded, changes[0].Kind)
	assert.Equal(t, testRuleNoUnusedVars, changes[0].Path)
}

func TestDiffRemoved(t *testing.T) {
	t.Parallel()

	before := &config.OxlintConfig{
		Rules: map[string]any{testRuleNoUnusedVars: testSeverityError},
	}
	after := &config.OxlintConfig{Rules: map[string]any{}}

	d := NewDiffer(before, after)
	changes := d.Diff()
	assert.Len(t, changes, 1)
	assert.Equal(t, KindRemoved, changes[0].Kind)
}

func TestDiffChanged(t *testing.T) {
	t.Parallel()

	before := &config.OxlintConfig{Rules: map[string]any{testRuleNoUnusedVars: testSeverityWarn}}
	after := &config.OxlintConfig{Rules: map[string]any{testRuleNoUnusedVars: testSeverityError}}

	d := NewDiffer(before, after)
	changes := d.Diff()
	assert.Len(t, changes, 1)
	assert.Equal(t, KindModified, changes[0].Kind)
	assert.Equal(t, testSeverityWarn, changes[0].Old)
	assert.Equal(t, testSeverityError, changes[0].New)
}

func TestDiffSummary(t *testing.T) {
	t.Parallel()

	before := &config.OxlintConfig{Rules: map[string]any{"a": testSeverityWarn}}
	after := &config.OxlintConfig{
		Rules: map[string]any{"a": testSeverityError, "b": testSeverityError},
	}

	d := NewDiffer(before, after)
	summary := d.Summary()
	assert.Contains(t, summary, "Added: 1")
	assert.Contains(t, summary, "Modified: 1")
}

func TestDiffPluginsAdded(t *testing.T) {
	t.Parallel()

	before := &config.OxlintConfig{Plugins: []string{testPluginTS}}
	after := &config.OxlintConfig{Plugins: []string{testPluginTS, testPluginReact, testPluginVue}}

	d := NewDiffer(before, after)
	changes := d.Diff()
	addedPlugins := filterChanges(changes, "plugin:")
	assert.Len(t, addedPlugins, 2)

	for _, c := range addedPlugins {
		assert.Equal(t, KindAdded, c.Kind)
	}
}

func TestDiffPluginsRemoved(t *testing.T) {
	t.Parallel()

	before := &config.OxlintConfig{Plugins: []string{testPluginTS, testPluginReact, testPluginVue}}
	after := &config.OxlintConfig{Plugins: []string{testPluginTS}}

	d := NewDiffer(before, after)
	changes := d.Diff()
	removedPlugins := filterChanges(changes, "plugin:")
	assert.Len(t, removedPlugins, 2)

	for _, c := range removedPlugins {
		assert.Equal(t, KindRemoved, c.Kind)
	}
}

func TestDiffEnvChanged(t *testing.T) {
	t.Parallel()

	before := &config.OxlintConfig{Env: map[string]bool{"builtin": true, "browser": false}}
	after := &config.OxlintConfig{Env: map[string]bool{"builtin": true, "node": true}}

	d := NewDiffer(before, after)
	changes := d.Diff()
	envChanges := filterChanges(changes, "env:")
	assert.Len(t, envChanges, 2) // browser removed, node added
}

func TestDiffSettingsChanged(t *testing.T) {
	t.Parallel()

	before := &config.OxlintConfig{Settings: map[string]any{
		"react": map[string]any{"version": "detect"},
	}}
	after := &config.OxlintConfig{Settings: map[string]any{
		"react":    map[string]any{"version": "18.0"},
		"jsx-a11y": map[string]any{"components": map[string]any{}},
	}}

	d := NewDiffer(before, after)
	changes := d.Diff()
	settingsChanges := filterChanges(changes, "settings:")
	assert.Len(t, settingsChanges, 2) // react changed, jsx-a11y added
}

func filterChanges(changes []Change, prefix string) []Change {
	var filtered []Change

	for _, c := range changes {
		if len(c.Path) >= len(prefix) && c.Path[:len(prefix)] == prefix {
			filtered = append(filtered, c)
		}
	}

	return filtered
}

func TestDiffFormatDiff(t *testing.T) {
	t.Parallel()

	before := &config.OxlintConfig{Rules: map[string]any{"a": testSeverityWarn}}
	after := &config.OxlintConfig{
		Rules: map[string]any{"a": testSeverityError, "b": testSeverityError},
	}

	d := NewDiffer(before, after)
	output := d.FormatDiff()
	assert.Contains(t, output, "~ a: warn → error")
	assert.Contains(t, output, "+ b: error")
}

func TestDiffJsPluginsAdded(t *testing.T) {
	t.Parallel()

	before := &config.OxlintConfig{}
	after := &config.OxlintConfig{JsPlugins: []string{"@shadcn/lint"}}

	d := NewDiffer(before, after)
	changes := d.Diff()
	jsPluginChanges := filterChanges(changes, "jsPlugin:")
	assert.Len(t, jsPluginChanges, 1)
	assert.Equal(t, KindAdded, jsPluginChanges[0].Kind)
	assert.Equal(t, "jsPlugin:@shadcn/lint", jsPluginChanges[0].Path)
}

func TestDiffJsPluginsRemoved(t *testing.T) {
	t.Parallel()

	before := &config.OxlintConfig{JsPlugins: []string{"@shadcn/lint"}}
	after := &config.OxlintConfig{}

	d := NewDiffer(before, after)
	changes := d.Diff()
	jsPluginChanges := filterChanges(changes, "jsPlugin:")
	assert.Len(t, jsPluginChanges, 1)
	assert.Equal(t, KindRemoved, jsPluginChanges[0].Kind)
}

func TestDiffRulesWithOptionArrays(t *testing.T) {
	t.Parallel()

	oldOptions := []any{"error", map[string]any{"allow": []any{"layout"}}}
	newOptions := []any{"error", map[string]any{"allow": []any{"layout", "spacing"}}}

	before := &config.OxlintConfig{Rules: map[string]any{"shadcn/no-restyle": oldOptions}}
	after := &config.OxlintConfig{Rules: map[string]any{"shadcn/no-restyle": newOptions}}

	d := NewDiffer(before, after)
	changes := d.Diff()
	assert.Len(t, changes, 1)
	assert.Equal(t, KindModified, changes[0].Kind)
	assert.Equal(t, `["error",{"allow":["layout"]}]`, changes[0].Old)
	assert.Equal(t, `["error",{"allow":["layout","spacing"]}]`, changes[0].New)
}
