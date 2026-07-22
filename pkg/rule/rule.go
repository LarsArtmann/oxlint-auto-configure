// Package rule defines types and the registry for oxlint rules.
package rule

import "slices"

// Category represents an oxlint rule category.
type Category string

// CategoryCorrectness and other category constants classify oxlint rules by concern.
const (
	CategoryCorrectness Category = "correctness" // correctness issues are bugs
	CategorySuspicious  Category = "suspicious"  // suspicious code patterns
	CategoryPedantic    Category = "pedantic"    // pedantic style rules
	CategoryPerf        Category = "perf"        // performance issues
	CategoryStyle       Category = "style"       // code style rules
	CategoryRestriction Category = "restriction" // restrictive patterns
	CategoryNursery     Category = "nursery"     // experimental rules
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

// PluginESLint and other plugin constants represent oxlint plugin sources.
const (
	PluginESLint     Plugin = "eslint"     // always-on default rules
	PluginImport     Plugin = "import"     // ES module import rules
	PluginJest       Plugin = "jest"       // Jest test framework rules
	PluginJSDoc      Plugin = "jsdoc"      // JSDoc annotation rules
	PluginJSXA11y    Plugin = "jsx_a11y"   // JSX accessibility rules
	PluginNextJS     Plugin = "nextjs"     // Next.js framework rules
	PluginNode       Plugin = "node"       // Node.js runtime rules
	PluginOXC        Plugin = "oxc"        // always-on OXC rules
	PluginPromise    Plugin = "promise"    // Promise/async patterns
	PluginReact      Plugin = "react"      // React component rules
	PluginReactPerf  Plugin = "react_perf" // React performance rules
	PluginTypeScript Plugin = "typescript" // always-on TypeScript rules
	PluginUnicorn    Plugin = "unicorn"    // always-on Unicorn rules
	PluginVitest     Plugin = "vitest"     // Vitest test framework rules
	PluginVue        Plugin = "vue"        // Vue component rules
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

// alwaysOnPlugins are enabled by oxlint without any CLI flag.
var alwaysOnPlugins = map[Plugin]bool{ //nolint:gochecknoglobals // immutable lookup table
	PluginESLint:     true,
	PluginUnicorn:    true,
	PluginTypeScript: true,
	PluginOXC:        true,
}

// cliFlagMap maps plugins to their oxlint CLI flag.
var cliFlagMap = map[Plugin]string{ //nolint:gochecknoglobals // immutable lookup table
	PluginJSXA11y:   "--jsx-a11y-plugin",
	PluginReactPerf: "--react-perf-plugin",
	PluginImport:    "--import-plugin",
	PluginJest:      "--jest-plugin",
	PluginJSDoc:     "--jsdoc-plugin",
	PluginNextJS:    "--nextjs-plugin",
	PluginNode:      "--node-plugin",
	PluginPromise:   "--promise-plugin",
	PluginReact:     "--react-plugin",
	PluginVitest:    "--vitest-plugin",
	PluginVue:       "--vue-plugin",
}

// NeedsFlag returns true if this plugin requires an explicit CLI flag to enable.
func (p Plugin) NeedsFlag() bool {
	return !alwaysOnPlugins[p]
}

// CLIFlag returns the oxlint CLI flag to enable this plugin (e.g., "--react-plugin").
// Returns empty string for always-on plugins (eslint, unicorn, typescript, oxc).
func (p Plugin) CLIFlag() string {
	if flag, ok := cliFlagMap[p]; ok {
		return flag
	}

	if !p.NeedsFlag() {
		return ""
	}

	return "--" + string(p) + "-plugin"
}

// FixCapability indicates what kind of auto-fix a rule supports.
type FixCapability string

// FixNone and other capability constants indicate what auto-fix support a rule has.
const (
	FixNone       FixCapability = "none"       // no auto-fix available
	FixSafe       FixCapability = "safe"       // safe auto-fix
	FixSuggestion FixCapability = "suggestion" // suggested fix
	FixDangerous  FixCapability = "dangerous"  // dangerous auto-fix
)

// Rule represents a single oxlint rule with all its metadata.
type Rule struct {
	Name      string        // e.g., "no-unused-vars"
	Plugin    Plugin        // e.g., "eslint"
	Category  Category      // e.g., "correctness"
	Enabled   bool          // enabled by default?
	TypeAware bool          // requires type information?
	Fix       FixCapability // auto-fix capability
	DocsURL   string        // documentation URL
}

// FullName returns the fully qualified rule name with plugin prefix.
// Matches oxlint's config key format: "typescript/no-floating-promises".
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

// SeverityError and other severity constants represent rule violation outcomes.
const (
	SeverityError SeverityDecision = "error" // rule violation is an error
	SeverityWarn  SeverityDecision = "warn"  // rule violation is a warning
	SeverityOff   SeverityDecision = "off"   // rule is disabled
)

// String returns the string representation.
func (s SeverityDecision) String() string { return string(s) }

// IsValid returns true if the severity is a recognized value.
func (s SeverityDecision) IsValid() bool {
	switch s {
	case SeverityError, SeverityWarn, SeverityOff:
		return true
	}

	return false
}
