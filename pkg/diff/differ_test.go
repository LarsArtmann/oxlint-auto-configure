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
		Rules: map[string]string{testRuleNoUnusedVars: testSeverityError},
	}
	after := &config.OxlintConfig{Rules: map[string]string{testRuleNoUnusedVars: testSeverityError}}

	d := NewDiffer(before, after)
	changes := d.Diff()
	assert.Empty(t, changes)
}

func TestDiffAdded(t *testing.T) {
	t.Parallel()

	before := &config.OxlintConfig{Rules: map[string]string{}}
	after := &config.OxlintConfig{Rules: map[string]string{testRuleNoUnusedVars: testSeverityError}}

	d := NewDiffer(before, after)
	changes := d.Diff()
	assert.Len(t, changes, 1)
	assert.Equal(t, KindAdded, changes[0].Kind)
	assert.Equal(t, testRuleNoUnusedVars, changes[0].Rule)
}

func TestDiffRemoved(t *testing.T) {
	t.Parallel()

	before := &config.OxlintConfig{
		Rules: map[string]string{testRuleNoUnusedVars: testSeverityError},
	}
	after := &config.OxlintConfig{Rules: map[string]string{}}

	d := NewDiffer(before, after)
	changes := d.Diff()
	assert.Len(t, changes, 1)
	assert.Equal(t, KindRemoved, changes[0].Kind)
}

func TestDiffChanged(t *testing.T) {
	t.Parallel()

	before := &config.OxlintConfig{Rules: map[string]string{testRuleNoUnusedVars: testSeverityWarn}}
	after := &config.OxlintConfig{Rules: map[string]string{testRuleNoUnusedVars: testSeverityError}}

	d := NewDiffer(before, after)
	changes := d.Diff()
	assert.Len(t, changes, 1)
	assert.Equal(t, KindChanged, changes[0].Kind)
	assert.Equal(t, testSeverityWarn, changes[0].OldValue)
	assert.Equal(t, testSeverityError, changes[0].NewValue)
}

func TestDiffSummary(t *testing.T) {
	t.Parallel()

	before := &config.OxlintConfig{Rules: map[string]string{"a": testSeverityWarn}}
	after := &config.OxlintConfig{
		Rules: map[string]string{"a": testSeverityError, "b": testSeverityError},
	}

	d := NewDiffer(before, after)
	summary := d.Summary()
	assert.Contains(t, summary, "Added: 1")
	assert.Contains(t, summary, "Changed: 1")
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
		if len(c.Rule) >= len(prefix) && c.Rule[:len(prefix)] == prefix {
			filtered = append(filtered, c)
		}
	}

	return filtered
}

func TestDiffFormatDiff(t *testing.T) {
	t.Parallel()

	before := &config.OxlintConfig{Rules: map[string]string{"a": testSeverityWarn}}
	after := &config.OxlintConfig{
		Rules: map[string]string{"a": testSeverityError, "b": testSeverityError},
	}

	d := NewDiffer(before, after)
	output := d.FormatDiff()
	assert.Contains(t, output, "~ a: warn → error")
	assert.Contains(t, output, "+ b: error")
}
