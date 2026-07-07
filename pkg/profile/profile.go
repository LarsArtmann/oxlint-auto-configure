// Package profile defines configuration profiles and the categorization engine
// that maps oxlint rule categories to severity decisions.
package profile

import (
	"fmt"
	"slices"
	"strings"

	"github.com/larsartmann/oxlint-auto-configure/pkg/rule"
)

// Profile determines how rules are categorized by severity.
type Profile string

const (
	// ProfileMaximalTypesafe enables EVERY rule at "error".
	// Maximum type safety and correctness enforcement.
	ProfileMaximalTypesafe Profile = "maximal-typesafe"

	// ProfileRecommended enables correctness+suspicious at error,
	// typescript+oxc at error, perf+style+pedantic at warn,
	// restriction at warn, nursery at off.
	ProfileRecommended Profile = "recommended"

	// ProfileStrict enables correctness+suspicious+typescript at error,
	// everything else at warn except nursery (off).
	ProfileStrict Profile = "strict"

	// ProfileMinimal keeps oxlint defaults and only enables
	// correctness at error.
	ProfileMinimal Profile = "minimal"
)

// AllProfiles returns all available profiles.
func AllProfiles() []Profile {
	return []Profile{ProfileMaximalTypesafe, ProfileRecommended, ProfileStrict, ProfileMinimal}
}

// AllProfileNames returns all profile names as strings.
func AllProfileNames() []string {
	ps := AllProfiles()
	names := make([]string, len(ps))
	for i, p := range ps {
		names[i] = string(p)
	}
	return names
}

// IsValid returns true if the profile is recognized.
func (p Profile) IsValid() bool {
	return slices.Contains(AllProfiles(), p)
}

// String returns the string representation.
func (p Profile) String() string { return string(p) }

// categoryPolicy defines the severity and config inclusion for a category.
type categoryPolicy struct {
	severity rule.SeverityDecision
	include  bool // false = omit from config (oxlint uses its defaults)
}

// profileSpec is the data-driven policy for a profile.
// explicit maps known categories to their policy.
// fallback is the policy for categories not in the map (e.g., unknown categories).
type profileSpec struct {
	explicit map[rule.Category]categoryPolicy
	fallback categoryPolicy
}

// profileSpecs is the single source of truth for all severity decisions.
// Adding a profile or category = adding one entry here.
var profileSpecs = map[Profile]profileSpec{ //nolint:gochecknoglobals // immutable policy table
	ProfileMaximalTypesafe: {
		explicit: map[rule.Category]categoryPolicy{
			rule.CategoryNursery: {rule.SeverityWarn, true},
		},
		fallback: categoryPolicy{rule.SeverityError, true},
	},
	ProfileRecommended: {
		explicit: map[rule.Category]categoryPolicy{
			rule.CategoryCorrectness: {rule.SeverityError, true},
			rule.CategorySuspicious:  {rule.SeverityError, true},
			rule.CategoryPerf:        {rule.SeverityWarn, true},
			rule.CategoryStyle:       {rule.SeverityWarn, true},
			rule.CategoryPedantic:    {rule.SeverityWarn, true},
			rule.CategoryRestriction: {rule.SeverityWarn, true},
			rule.CategoryNursery:     {rule.SeverityOff, true},
		},
		fallback: categoryPolicy{rule.SeverityOff, true},
	},
	ProfileStrict: {
		explicit: map[rule.Category]categoryPolicy{
			rule.CategoryCorrectness: {rule.SeverityError, true},
			rule.CategorySuspicious:  {rule.SeverityError, true},
			rule.CategoryNursery:     {rule.SeverityOff, true},
		},
		fallback: categoryPolicy{rule.SeverityWarn, true},
	},
	ProfileMinimal: {
		explicit: map[rule.Category]categoryPolicy{
			rule.CategoryCorrectness: {rule.SeverityError, true},
		},
		fallback: categoryPolicy{rule.SeverityOff, false},
	},
}

// Description returns a human-readable description derived from the profile's
// category severity decisions. One source of truth — cannot drift from decide logic.
func (p Profile) Description() string {
	if !p.IsValid() {
		return "unknown profile"
	}

	cat := NewCategorizer(p, nil)
	groups := make(map[string][]string) // severity → [category names]
	for _, c := range rule.AllCategories() {
		sev, include := cat.DecideCategory(c)
		if !include {
			groups["default"] = append(groups["default"], string(c))
			continue
		}
		key := string(sev)
		groups[key] = append(groups[key], string(c))
	}

	parts := make([]string, 0, len(groups))
	for _, sev := range []string{"error", "warn", "off", "default"} {
		cats, ok := groups[sev]
		if !ok {
			continue
		}
		slices.Sort(cats)
		parts = append(parts, fmt.Sprintf("%s at %s", strings.Join(cats, "+"), sev))
	}

	return strings.Join(parts, ", ")
}

// Categorizer maps rules to severity decisions based on a profile.
type Categorizer struct {
	profile      Profile
	pluginConfig PluginConfig
}

// PluginConfig determines which plugins are relevant for the target project.
type PluginConfig map[rule.Plugin]bool

// NewCategorizer creates a categorizer with the given profile and plugin config.
func NewCategorizer(p Profile, pc PluginConfig) *Categorizer {
	return &Categorizer{profile: p, pluginConfig: pc}
}

// DecideCategory returns the category-level severity for the given category.
// The bool indicates whether the category should be included in the config;
// false means omit it (oxlint will use its defaults).
func (c *Categorizer) DecideCategory(cat rule.Category) (rule.SeverityDecision, bool) {
	spec, ok := profileSpecs[c.profile]
	if !ok {
		return rule.SeverityOff, false
	}
	p, ok := spec.explicit[cat]
	if !ok {
		p = spec.fallback
	}
	return p.severity, p.include
}

// Decide returns the severity decision for a given rule.
// For the minimal profile, disabled rules are always off and enabled
// non-correctness rules get warn (left at oxlint default).
func (c *Categorizer) Decide(r rule.Rule) rule.SeverityDecision {
	if c.profile == ProfileMinimal {
		if !r.Enabled {
			return rule.SeverityOff
		}
		if r.Category == rule.CategoryCorrectness {
			return rule.SeverityError
		}
		return rule.SeverityWarn
	}
	sev, _ := c.DecideCategory(r.Category)
	return sev
}

// EnabledPlugins returns the list of detected plugins that are enabled.
func (c *Categorizer) EnabledPlugins() []rule.Plugin {
	plugins := make([]rule.Plugin, 0, len(c.pluginConfig))
	for p, enabled := range c.pluginConfig {
		if enabled {
			plugins = append(plugins, p)
		}
	}
	slices.Sort(plugins)
	return plugins
}

// IsPluginRelevant returns true if a rule's plugin is relevant given the project config.
func (c *Categorizer) IsPluginRelevant(r rule.Rule) bool {
	if !r.Plugin.NeedsFlag() {
		return true
	}
	return c.pluginConfig[r.Plugin]
}

// RuleDecision pairs a rule with its decided severity.
type RuleDecision struct {
	Rule     rule.Rule
	Severity rule.SeverityDecision
}

// DecideAll returns severity decisions for all rules in the registry.
func (c *Categorizer) DecideAll(reg *rule.Registry) []RuleDecision {
	rules := reg.All()
	decisions := make([]RuleDecision, 0, len(rules))
	for _, r := range rules {
		decisions = append(decisions, RuleDecision{
			Rule:     r,
			Severity: c.Decide(r),
		})
	}
	return decisions
}
