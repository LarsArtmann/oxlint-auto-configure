// Package diff compares two oxlint configs and shows the differences.
//
// The comparison engine lives in linter-autoconfigure-sdk (DiffMaps,
// DiffSets, DiffBlobs, Summary, FormatDiff); this package is the oxlint
// field projection: it maps OxlintConfig sections onto the engine's
// comparators and keeps the domain prefixes ("plugin:", "category:", ...).
package diff

import (
	"strconv"

	autoconfigure "github.com/larsartmann/linter-autoconfigure-sdk"
	"github.com/larsartmann/oxlint-auto-configure/pkg/config"
)

// Change is one difference between two configs: the SDK's autoconfigure.Change.
// Path carries the section-prefixed setting name (e.g. "plugin:import",
// "rules.no-console"); Kind is KindAdded, KindRemoved, or KindModified.
type Change = autoconfigure.Change

// Kind classifies a Change; alias of the SDK's Kind.
type Kind = autoconfigure.Kind

// Change-kind constants, re-exported from the SDK engine.
const (
	KindAdded    = autoconfigure.KindAdded    // rule/category was not present before
	KindRemoved  = autoconfigure.KindRemoved  // rule/category was present before but not after
	KindModified = autoconfigure.KindModified // rule/category value changed
)

// Differ compares two OxlintConfig instances.
type Differ struct {
	before *config.OxlintConfig
	after  *config.OxlintConfig
}

// NewDiffer creates a differ for two configs.
func NewDiffer(before, after *config.OxlintConfig) *Differ {
	return &Differ{before: before, after: after}
}

// Diff computes all changes between before and after configs, sorted by
// path (deterministic across runs).
func (d *Differ) Diff() []Change {
	changes := make([]Change, 0, 8)

	changes = append(changes, autoconfigure.DiffSets(d.before.Plugins, d.after.Plugins, "plugin:")...)
	changes = append(changes, autoconfigure.DiffSets(d.before.JsPlugins, d.after.JsPlugins, "jsPlugin:")...)
	changes = append(changes, autoconfigure.DiffMaps(d.before.Categories, d.after.Categories, "category:")...)
	changes = append(changes, autoconfigure.DiffMaps(stringify(d.before.Rules), stringify(d.after.Rules), "")...)
	changes = append(changes, autoconfigure.DiffMaps(formatBools(d.before.Env), formatBools(d.after.Env), "env:")...)
	changes = append(changes, autoconfigure.DiffMaps(stringify(d.before.Settings), stringify(d.after.Settings), "settings:")...)
	changes = append(changes, autoconfigure.DiffBlobs(canonicalBlocks(d.before.Overrides), canonicalBlocks(d.after.Overrides), "override:")...)

	return changes
}

// HasChanges reports whether the two configs differ at all.
func (d *Differ) HasChanges() bool {
	return len(d.Diff()) > 0
}

// Summary returns a one-line human-readable tally of the changes
// ("Added: 1, Modified: 2, Removed: 3").
func (d *Differ) Summary() string {
	return autoconfigure.Summary(d.Diff())
}

// FormatDiff returns a formatted string showing all changes: "+"/"-"/"~"
// lines sorted by path, "No changes." when the configs match.
func (d *Differ) FormatDiff() string {
	return autoconfigure.FormatDiff(d.Diff())
}

// stringify renders map[string]any values for comparison: plain strings
// (severities, settings strings) display bare; structured values (rule
// options, settings objects) display as deterministic compact JSON.
func stringify(m map[string]any) map[string]string {
	result := make(map[string]string, len(m))
	for key, value := range m {
		result[key] = autoconfigure.StringValue(value)
	}

	return result
}

// formatBools renders map[string]bool as strings for comparison.
func formatBools(m map[string]bool) map[string]string {
	result := make(map[string]string, len(m))
	for key, value := range m {
		result[key] = strconv.FormatBool(value)
	}

	return result
}

// canonicalBlocks renders each overrides block to its canonical form
// (deterministic compact JSON via the SDK's StringValue: sorted keys), so
// DiffBlobs compares them as an order-insensitive set.
func canonicalBlocks(blocks []map[string]any) []string {
	canonical := make([]string, 0, len(blocks))
	for _, block := range blocks {
		canonical = append(canonical, autoconfigure.StringValue(block))
	}

	return canonical
}
