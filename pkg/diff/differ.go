// Package diff compares two oxlint configs and shows the differences.
package diff

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/larsartmann/oxlint-auto-configure/pkg/config"
)

// Change represents a single difference between two configs.
type Change struct {
	Rule     string
	OldValue string
	NewValue string
	Kind     ChangeKind
}

// ChangeKind categorizes the type of change.
type ChangeKind string

// KindAdded and other change kind constants categorize config differences.
const (
	KindAdded     ChangeKind = "added"     // rule/category was not present before
	KindRemoved   ChangeKind = "removed"   // rule/category was present before but not after
	KindChanged   ChangeKind = "changed"   // rule/category value changed
	KindUnchanged ChangeKind = "unchanged" // rule/category value stayed the same
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

// Diff computes all changes between before and after configs.
func (d *Differ) Diff() []Change {
	changes := make([]Change, 0, 8)

	changes = append(changes, d.compareSlices(d.before.Plugins, d.after.Plugins, "plugin:")...)
	changes = append(
		changes,
		d.compareMaps(d.before.Categories, d.after.Categories, "category:")...,
	)
	changes = append(changes, d.compareMaps(d.before.Rules, d.after.Rules, "")...)
	changes = append(changes, d.compareBoolMaps(d.before.Env, d.after.Env, "env:")...)
	changes = append(changes, d.compareAnyMaps(d.before.Settings, d.after.Settings, "settings:")...)

	return changes
}

// compareMaps returns changes between two string maps with an optional prefix.
func (d *Differ) compareMaps(before, after map[string]string, prefix string) []Change {
	var changes []Change

	for _, key := range d.collectAllKeys(before, after) {
		bv, hadBefore := before[key]
		av, hasAfter := after[key]
		name := prefix + key

		switch {
		case !hadBefore && hasAfter:
			changes = append(
				changes,
				Change{Rule: name, OldValue: "", NewValue: av, Kind: KindAdded},
			)
		case hadBefore && !hasAfter:
			changes = append(
				changes,
				Change{Rule: name, OldValue: bv, NewValue: "", Kind: KindRemoved},
			)
		case hadBefore && hasAfter && bv != av:
			changes = append(
				changes,
				Change{Rule: name, OldValue: bv, NewValue: av, Kind: KindChanged},
			)
		}
	}

	return changes
}

// Summary returns a human-readable summary of changes.
func (d *Differ) Summary() string {
	changes := d.Diff()

	added, removed, changed := 0, 0, 0
	for _, c := range changes {
		switch c.Kind {
		case KindAdded:
			added++
		case KindRemoved:
			removed++
		case KindChanged:
			changed++
		case KindUnchanged:
			// no-op
		}
	}

	return fmt.Sprintf("Added: %d, Changed: %d, Removed: %d", added, changed, removed)
}

// FormatDiff returns a formatted string showing all changes.
func (d *Differ) FormatDiff() string {
	changes := d.Diff()
	if len(changes) == 0 {
		return "No changes."
	}

	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Rule < changes[j].Rule
	})

	var b strings.Builder
	for _, c := range changes {
		switch c.Kind {
		case KindAdded:
			fmt.Fprintf(&b, "+ %s: %s\n", c.Rule, c.NewValue)
		case KindRemoved:
			fmt.Fprintf(&b, "- %s: %s\n", c.Rule, c.OldValue)
		case KindChanged:
			fmt.Fprintf(&b, "~ %s: %s → %s\n", c.Rule, c.OldValue, c.NewValue)
		case KindUnchanged:
			// no-op
		}
	}

	return b.String()
}

// compareSlices compares two string slices for added/removed items.
func (d *Differ) compareSlices(before, after []string, prefix string) []Change {
	beforeSet := make(map[string]bool, len(before))
	for _, s := range before {
		beforeSet[s] = true
	}
	afterSet := make(map[string]bool, len(after))
	for _, s := range after {
		afterSet[s] = true
	}

	var changes []Change
	seen := make(map[string]bool)

	for _, s := range before {
		if seen[s] {
			continue
		}
		seen[s] = true
		if !afterSet[s] {
			changes = append(
				changes,
				Change{Rule: prefix + s, OldValue: s, NewValue: "", Kind: KindRemoved},
			)
		}
	}
	for _, s := range after {
		if seen[s] {
			continue
		}
		seen[s] = true
		if !beforeSet[s] {
			changes = append(
				changes,
				Change{Rule: prefix + s, OldValue: "", NewValue: s, Kind: KindAdded},
			)
		}
	}
	return changes
}

// compareBoolMaps compares two map[string]bool by converting to string values.
func (d *Differ) compareBoolMaps(before, after map[string]bool, prefix string) []Change {
	bStr := make(map[string]string, len(before))
	for k, v := range before {
		bStr[k] = strconv.FormatBool(v)
	}
	aStr := make(map[string]string, len(after))
	for k, v := range after {
		aStr[k] = strconv.FormatBool(v)
	}
	return d.compareMaps(bStr, aStr, prefix)
}

// compareAnyMaps compares two map[string]any by JSON-serializing values for comparison.
func (d *Differ) compareAnyMaps(before, after map[string]any, prefix string) []Change {
	stringify := func(m map[string]any) map[string]string {
		result := make(map[string]string, len(m))
		for k, v := range m {
			data, err := json.Marshal(v)
			if err != nil {
				result[k] = fmt.Sprintf("%v", v)
				continue
			}
			result[k] = string(data)
		}
		return result
	}
	return d.compareMaps(stringify(before), stringify(after), prefix)
}

func (d *Differ) collectAllKeys(before, after map[string]string) []string {
	seen := make(map[string]struct{})
	var keys []string

	addUnseenKeys(seen, &keys, before)
	addUnseenKeys(seen, &keys, after)

	return keys
}

func addUnseenKeys(seen map[string]struct{}, keys *[]string, m map[string]string) {
	for k := range m {
		if _, exists := seen[k]; !exists {
			seen[k] = struct{}{}
			*keys = append(*keys, k)
		}
	}
}
