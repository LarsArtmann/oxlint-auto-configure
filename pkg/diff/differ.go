// Package diff compares two oxlint configs and shows the differences.
package diff

import (
	"fmt"
	"sort"
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

const (
	KindAdded      ChangeKind = "added"
	KindRemoved    ChangeKind = "removed"
	KindChanged    ChangeKind = "changed"
	KindUnchanged  ChangeKind = "unchanged"
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
	var changes []Change

	for _, key := range d.collectAllKeys(d.before.Categories, d.after.Categories) {
		before, hadBefore := d.before.Categories[key]
		after, hasAfter := d.after.Categories[key]

		switch {
		case !hadBefore && hasAfter:
			changes = append(changes, Change{Rule: "category:" + key, OldValue: "", NewValue: after, Kind: KindAdded})
		case hadBefore && !hasAfter:
			changes = append(changes, Change{Rule: "category:" + key, OldValue: before, NewValue: "", Kind: KindRemoved})
		case hadBefore && hasAfter && before != after:
			changes = append(changes, Change{Rule: "category:" + key, OldValue: before, NewValue: after, Kind: KindChanged})
		}
	}

	for _, rule := range d.collectAllKeys(d.before.Rules, d.after.Rules) {
		before, hadBefore := d.before.Rules[rule]
		after, hasAfter := d.after.Rules[rule]

		switch {
		case !hadBefore && hasAfter:
			changes = append(changes, Change{Rule: rule, OldValue: "", NewValue: after, Kind: KindAdded})
		case hadBefore && !hasAfter:
			changes = append(changes, Change{Rule: rule, OldValue: before, NewValue: "", Kind: KindRemoved})
		case hadBefore && hasAfter && before != after:
			changes = append(changes, Change{Rule: rule, OldValue: before, NewValue: after, Kind: KindChanged})
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
		}
	}

	return b.String()
}

func (d *Differ) collectAllKeys(before, after map[string]string) []string {
	seen := make(map[string]struct{})
	var keys []string

	for k := range before {
		if _, exists := seen[k]; !exists {
			seen[k] = struct{}{}
			keys = append(keys, k)
		}
	}
	for k := range after {
		if _, exists := seen[k]; !exists {
			seen[k] = struct{}{}
			keys = append(keys, k)
		}
	}

	return keys
}
