// Package profile defines configuration profiles and the categorization engine
// that maps oxlint rule categories to severity decisions.
package profile

import (
	"slices"
	"sort"

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

// IsValid returns true if the profile is recognized.
func (p Profile) IsValid() bool {
	return slices.Contains(AllProfiles(), p)
}

// String returns the string representation.
func (p Profile) String() string { return string(p) }

// Description returns a human-readable description of the profile.
func (p Profile) Description() string {
	switch p {
	case ProfileMaximalTypesafe:
		return "Enable ALL rules at 'error' — maximum type safety and correctness enforcement"
	case ProfileRecommended:
		return "Correctness+suspicious+TypeScript at error, perf+style+pedantic at warn, restriction at warn, nursery at off"
	case ProfileStrict:
		return "Correctness+suspicious+TypeScript+OXC at error, everything else at warn except nursery (off)"
	case ProfileMinimal:
		return "Only correctness at error, everything else uses oxlint defaults"
	default:
		return "unknown profile"
	}
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

// Decide returns the severity decision for a given rule.
func (c *Categorizer) Decide(r rule.Rule) rule.SeverityDecision {
	switch c.profile {
	case ProfileMaximalTypesafe:
		return c.decideMaximal(r)
	case ProfileRecommended:
		return c.decideRecommended(r)
	case ProfileStrict:
		return c.decideStrict(r)
	case ProfileMinimal:
		return c.decideMinimal(r)
	default:
		return rule.SeverityOff
	}
}

// EnabledPlugins returns the list of detected plugins that are enabled.
func (c *Categorizer) EnabledPlugins() []rule.Plugin {
	plugins := make([]rule.Plugin, 0, len(c.pluginConfig))
	for p, enabled := range c.pluginConfig {
		if enabled {
			plugins = append(plugins, p)
		}
	}
	sort.Slice(plugins, func(i, j int) bool {
		return string(plugins[i]) < string(plugins[j])
	})
	return plugins
}

// decideMaximal: ALL rules at error. Maximum type safety.
func (c *Categorizer) decideMaximal(r rule.Rule) rule.SeverityDecision {
	if r.Category == rule.CategoryNursery {
		return rule.SeverityWarn
	}
	return rule.SeverityError
}

// decideRecommended: correctness+suspicious+TS at error, perf/style/pedantic at warn,
// restriction at warn, nursery at off.
func (c *Categorizer) decideRecommended(r rule.Rule) rule.SeverityDecision {
	switch r.Category {
	case rule.CategoryCorrectness, rule.CategorySuspicious:
		return rule.SeverityError
	case rule.CategoryPerf, rule.CategoryStyle, rule.CategoryPedantic:
		return rule.SeverityWarn
	case rule.CategoryRestriction:
		return rule.SeverityWarn
	case rule.CategoryNursery:
		return rule.SeverityOff
	default:
		return rule.SeverityOff
	}
}

// decideStrict: correctness+suspicious+TS+OXC at error, everything else at warn except nursery.
func (c *Categorizer) decideStrict(r rule.Rule) rule.SeverityDecision {
	if r.Category == rule.CategoryNursery {
		return rule.SeverityOff
	}

	switch r.Category {
	case rule.CategoryCorrectness, rule.CategorySuspicious:
		return rule.SeverityError
	default:
		return rule.SeverityWarn
	}
}

// decideMinimal: only correctness at error, everything else uses oxlint defaults.
func (c *Categorizer) decideMinimal(r rule.Rule) rule.SeverityDecision {
	if r.Category == rule.CategoryCorrectness && r.Enabled {
		return rule.SeverityError
	}
	if r.Enabled {
		return rule.SeverityWarn
	}
	return rule.SeverityOff
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
