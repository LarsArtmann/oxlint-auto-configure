// Package rule defines types and the registry for oxlint rules.
package rule

import "slices"

// Category represents an oxlint rule category.
type Category string

const (
	CategoryCorrectness Category = "correctness"
	CategorySuspicious  Category = "suspicious"
	CategoryPedantic    Category = "pedantic"
	CategoryPerf        Category = "perf"
	CategoryStyle       Category = "style"
	CategoryRestriction Category = "restriction"
	CategoryNursery     Category = "nursery"
)

// AllCategories returns all known oxlint rule categories.
func AllCategories() []Category {
	return []Category{
		CategoryCorrectness,
		CategorySuspicious,
		CategoryPedantic,
		CategoryPerf,
		CategoryStyle,
		CategoryRestriction,
		CategoryNursery,
	}
}

// IsValid returns true if the category is a recognized oxlint category.
func (c Category) IsValid() bool {
	switch c {
	case CategoryCorrectness, CategorySuspicious, CategoryPedantic,
		CategoryPerf, CategoryStyle, CategoryRestriction, CategoryNursery:
		return true
	}
	return false
}

// String returns the string representation.
func (c Category) String() string { return string(c) }

// Plugin represents an oxlint plugin source.
type Plugin string

const (
	PluginESLint    Plugin = "eslint"
	PluginImport    Plugin = "import"
	PluginJest      Plugin = "jest"
	PluginJSDoc     Plugin = "jsdoc"
	PluginJSXA11y   Plugin = "jsx_a11y"
	PluginNextJS    Plugin = "nextjs"
	PluginNode      Plugin = "node"
	PluginOXC       Plugin = "oxc"
	PluginPromise   Plugin = "promise"
	PluginReact     Plugin = "react"
	PluginReactPerf Plugin = "react_perf"
	PluginTypeScript Plugin = "typescript"
	PluginUnicorn   Plugin = "unicorn"
	PluginVitest    Plugin = "vitest"
	PluginVue       Plugin = "vue"
)

// AllPlugins returns all known oxlint plugins.
func AllPlugins() []Plugin {
	return []Plugin{
		PluginESLint, PluginImport, PluginJest, PluginJSDoc,
		PluginJSXA11y, PluginNextJS, PluginNode, PluginOXC,
		PluginPromise, PluginReact, PluginReactPerf, PluginTypeScript,
		PluginUnicorn, PluginVitest, PluginVue,
	}
}

// IsValid returns true if the plugin is a recognized oxlint plugin.
func (p Plugin) IsValid() bool {
	return slices.Contains(AllPlugins(), p)
}

// String returns the string representation.
func (p Plugin) String() string { return string(p) }

// FixCapability indicates what kind of auto-fix a rule supports.
type FixCapability string

const (
	FixNone       FixCapability = "none"
	FixSafe       FixCapability = "safe"
	FixSuggestion FixCapability = "suggestion"
	FixDangerous  FixCapability = "dangerous"
)

// Rule represents a single oxlint rule with all its metadata.
type Rule struct {
	Name        string        // e.g., "no-unused-vars"
	Plugin      Plugin        // e.g., "eslint"
	Category    Category      // e.g., "correctness"
	Enabled     bool          // enabled by default?
	TypeAware   bool          // requires type information?
	Fix         FixCapability // auto-fix capability
	DocsURL     string        // documentation URL
}

// FullName returns the fully qualified rule name with plugin prefix.
// Matches oxlint's config key format: "typescript/no-floating-promises"
func (r Rule) FullName() string {
	if r.Plugin == PluginESLint {
		return r.Name
	}
	return string(r.Plugin) + "/" + r.Name
}

// IsFixable returns true if the rule has any auto-fix capability.
func (r Rule) IsFixable() bool {
	return r.Fix != FixNone
}

// IsSafeFixable returns true if the rule has a safe auto-fix.
func (r Rule) IsSafeFixable() bool {
	return r.Fix == FixSafe
}

// SeverityDecision represents the chosen severity for a rule.
type SeverityDecision string

const (
	SeverityError   SeverityDecision = "error"
	SeverityWarn    SeverityDecision = "warn"
	SeverityOff     SeverityDecision = "off"
)

// String returns the string representation.
func (s SeverityDecision) String() string { return string(s) }
