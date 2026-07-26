package rule

import (
	"embed"
	"encoding/json/v2"
	"fmt"
	"slices"
	"sort"
	"strings"
)

//go:embed rules_data.json
var rulesDataFS embed.FS

//go:embed rules_version.txt
var rulesVersionData string

// EmbeddedVersion returns the oxlint version that generated the embedded rules.
func EmbeddedVersion() string {
	return strings.TrimSpace(rulesVersionData)
}

// rawRule maps the oxlint JSON rule format.
type rawRule struct {
	Scope     string `json:"scope"`
	Value     string `json:"value"`
	Category  string `json:"category"`
	TypeAware bool   `json:"type_aware"`
	Fix       string `json:"fix"`
	Default   bool   `json:"default"`
	DocsURL   string `json:"docs_url"`
}

// Registry holds all known oxlint rules, indexed for fast lookup.
type Registry struct {
	rules    []Rule
	byName   map[string]Rule
	byPlugin map[Plugin][]Rule
	byCat    map[Category][]Rule
}

// LoadRegistry loads all rules from the embedded JSON data.
func LoadRegistry() (*Registry, error) {
	data, err := rulesDataFS.ReadFile("rules_data.json")
	if err != nil {
		return nil, fmt.Errorf("read embedded rules data: %w", err)
	}

	var raw []rawRule
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse rules JSON: %w", err)
	}

	reg := &Registry{
		rules:    make([]Rule, 0, len(raw)),
		byName:   make(map[string]Rule, len(raw)),
		byPlugin: make(map[Plugin][]Rule),
		byCat:    make(map[Category][]Rule),
	}

	for _, r := range raw {
		rule := Rule{
			Name:      r.Value,
			Plugin:    Plugin(r.Scope),
			Category:  Category(r.Category),
			Enabled:   r.Default,
			TypeAware: r.TypeAware,
			Fix:       mapFix(r.Fix),
			DocsURL:   r.DocsURL,
		}

		if !rule.Category.IsValid() || !rule.Plugin.IsValid() {
			continue
		}

		reg.rules = append(reg.rules, rule)
		reg.byName[rule.FullName()] = rule
		reg.byPlugin[rule.Plugin] = append(reg.byPlugin[rule.Plugin], rule)
		reg.byCat[rule.Category] = append(reg.byCat[rule.Category], rule)
	}

	sort.Slice(reg.rules, func(i, j int) bool {
		if reg.rules[i].Category != reg.rules[j].Category {
			return reg.rules[i].Category < reg.rules[j].Category
		}

		return reg.rules[i].FullName() < reg.rules[j].FullName()
	})

	return reg, nil
}

// All returns all rules sorted by category then name.
func (r *Registry) All() []Rule {
	result := slices.Clone(r.rules)

	return result
}

// ByName looks up a rule by its fully qualified name.
func (r *Registry) ByName(name string) (Rule, bool) {
	rule, ok := r.byName[name]

	return rule, ok
}

// ByPlugin returns all rules for a given plugin.
func (r *Registry) ByPlugin(p Plugin) []Rule {
	return r.byPlugin[p]
}

// ByCategory returns all rules for a given category.
func (r *Registry) ByCategory(c Category) []Rule {
	return r.byCat[c]
}

// Filter returns all rules matching the given predicate.
func (r *Registry) Filter(pred func(Rule) bool) []Rule {
	var result []Rule

	for _, rule := range r.rules {
		if pred(rule) {
			result = append(result, rule)
		}
	}

	return result
}

// EnabledByDefault returns all rules that are enabled by default in oxlint.
func (r *Registry) EnabledByDefault() []Rule {
	return r.Filter(func(rule Rule) bool { return rule.Enabled })
}

// DisabledByDefault returns all rules that are disabled by default in oxlint.
func (r *Registry) DisabledByDefault() []Rule {
	return r.Filter(func(rule Rule) bool { return !rule.Enabled })
}

// Fixable returns all rules that have any fix capability.
func (r *Registry) Fixable() []Rule {
	return r.Filter(func(rule Rule) bool { return rule.IsFixable() })
}

// TypeAwareRules returns all rules that require type information.
func (r *Registry) TypeAwareRules() []Rule {
	return r.Filter(func(rule Rule) bool { return rule.TypeAware })
}

// Len returns the total number of rules.
func (r *Registry) Len() int {
	return len(r.rules)
}

// mapFix converts the oxlint fix string to a FixCapability.
func mapFix(fix string) FixCapability {
	switch fix {
	case "none", "pending":
		return FixNone
	case "fixable_fix", "fixable_safe_fix_or_suggestion", "conditional_safe_fix_or_suggestion":
		return FixSafe
	case "fixable_suggestion", "conditional_suggestion":
		return FixSuggestion
	case "fixable_dangerous_fix", "fixable_dangerous_suggestion",
		"fixable_dangerous_fix_or_suggestion", "conditional_dangerous_fix",
		"conditional_dangerous_fix_or_suggestion", "conditional_fix":
		return FixDangerous
	default:
		return FixNone
	}
}
