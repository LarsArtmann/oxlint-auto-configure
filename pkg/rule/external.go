package rule

import (
	"slices"
	"strings"
)

// ExternalPlugin describes an oxlint JS plugin loaded at runtime from an npm
// package via the "jsPlugins" config key, rather than a built-in plugin.
// Unlike built-in plugins, its rules are not part of the embedded registry:
// they are only known to the oxlint binary that loads the package.
type ExternalPlugin struct {
	// Package is the npm package name registered under "jsPlugins".
	Package string
	// Prefix is the rule-name prefix the plugin's rules carry (e.g. the
	// "shadcn" in "shadcn/no-restyle"). It doubles as the plugin's settings
	// key under "settings" (e.g. "settings.shadcn").
	Prefix string
}

// knownExternalPlugins lists the oxlint JS plugins this tool recognizes.
// Detection maps package.json dependencies to entries here; preservation and
// validation match rules and settings by Prefix.
//
//nolint:gochecknoglobals // immutable lookup table
var knownExternalPlugins = []ExternalPlugin{
	{
		// https://github.com/shadcn-ui/lint — agent-first design-system
		// linter for Tailwind v4 (React, Vue, Svelte). Requires oxlint 1.80+.
		Package: "@shadcn/lint",
		Prefix:  "shadcn",
	},
}

// KnownExternalPlugins returns the recognized oxlint JS plugins.
func KnownExternalPlugins() []ExternalPlugin {
	return slices.Clone(knownExternalPlugins)
}

// ExternalPluginByPackage returns the known external plugin for an npm
// package name, and whether it was found.
func ExternalPluginByPackage(pkg string) (ExternalPlugin, bool) {
	for _, p := range knownExternalPlugins {
		if p.Package == pkg {
			return p, true
		}
	}

	return ExternalPlugin{}, false
}

// ExternalPluginByRuleName returns the known external plugin owning a
// fully-qualified rule name (e.g. "shadcn/no-restyle"), and whether it was
// found.
func ExternalPluginByRuleName(name string) (ExternalPlugin, bool) {
	for _, p := range knownExternalPlugins {
		if hasPluginRulePrefix(name, p.Prefix) {
			return p, true
		}
	}

	return ExternalPlugin{}, false
}

// hasPluginRulePrefix reports whether name is a prefixed rule of plugin, in
// oxlint's "plugin/rule" key format (e.g. "shadcn/no-restyle"). Bare rule
// names never match: built-in eslint rules carry no prefix, and an unprefixed
// name cannot be attributed to an external plugin.
func hasPluginRulePrefix(name, plugin string) bool {
	prefix, rest, found := strings.Cut(name, "/")
	return found && prefix == plugin && rest != ""
}
