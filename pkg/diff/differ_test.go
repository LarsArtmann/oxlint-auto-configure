package diff

import (
	"testing"

	"github.com/larsartmann/oxlint-auto-configure/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestDiffNoChanges(t *testing.T) {
	before := &config.OxlintConfig{Rules: map[string]string{"no-unused-vars": "error"}}
	after := &config.OxlintConfig{Rules: map[string]string{"no-unused-vars": "error"}}

	d := NewDiffer(before, after)
	changes := d.Diff()
	assert.Empty(t, changes)
}

func TestDiffAdded(t *testing.T) {
	before := &config.OxlintConfig{Rules: map[string]string{}}
	after := &config.OxlintConfig{Rules: map[string]string{"no-unused-vars": "error"}}

	d := NewDiffer(before, after)
	changes := d.Diff()
	assert.Len(t, changes, 1)
	assert.Equal(t, KindAdded, changes[0].Kind)
	assert.Equal(t, "no-unused-vars", changes[0].Rule)
}

func TestDiffRemoved(t *testing.T) {
	before := &config.OxlintConfig{Rules: map[string]string{"no-unused-vars": "error"}}
	after := &config.OxlintConfig{Rules: map[string]string{}}

	d := NewDiffer(before, after)
	changes := d.Diff()
	assert.Len(t, changes, 1)
	assert.Equal(t, KindRemoved, changes[0].Kind)
}

func TestDiffChanged(t *testing.T) {
	before := &config.OxlintConfig{Rules: map[string]string{"no-unused-vars": "warn"}}
	after := &config.OxlintConfig{Rules: map[string]string{"no-unused-vars": "error"}}

	d := NewDiffer(before, after)
	changes := d.Diff()
	assert.Len(t, changes, 1)
	assert.Equal(t, KindChanged, changes[0].Kind)
	assert.Equal(t, "warn", changes[0].OldValue)
	assert.Equal(t, "error", changes[0].NewValue)
}

func TestDiffSummary(t *testing.T) {
	before := &config.OxlintConfig{Rules: map[string]string{"a": "warn"}}
	after := &config.OxlintConfig{Rules: map[string]string{"a": "error", "b": "error"}}

	d := NewDiffer(before, after)
	summary := d.Summary()
	assert.Contains(t, summary, "Added: 1")
	assert.Contains(t, summary, "Changed: 1")
}

func TestDiffFormatDiff(t *testing.T) {
	before := &config.OxlintConfig{Rules: map[string]string{"a": "warn"}}
	after := &config.OxlintConfig{Rules: map[string]string{"a": "error", "b": "error"}}

	d := NewDiffer(before, after)
	output := d.FormatDiff()
	assert.Contains(t, output, "~ a: warn → error")
	assert.Contains(t, output, "+ b: error")
}
