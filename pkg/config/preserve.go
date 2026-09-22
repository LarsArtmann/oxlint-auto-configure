package config

import (
	"slices"

	"github.com/larsartmann/oxlint-auto-configure/pkg/rule"
)

// PreserveExternal carries every external-plugin setting from an existing
// config into a freshly generated one, mutating and returning generated.
//
// Regeneration is destructive by design: the tool owns plugins, categories,
// rules, and env in the configs it writes. External JS plugins are the
// exception — their registrations, rule options, and settings cannot be
// derived from the embedded registry and represent deliberate setup
// (typically by an agent following @shadcn/lint's setup guide), so dropping
// them would be silent data loss. Everything an external plugin owns in the
// existing config survives a regeneration verbatim; entries are only ever
// added, never removed.
func PreserveExternal(existing, generated *OxlintConfig) *OxlintConfig {
	if existing == nil {
		return generated
	}

	preserveJsPlugins(existing, generated)
	preserveExternalRules(existing, generated)
	preserveExternalSettings(existing, generated)

	return generated
}

// preserveJsPlugins merges the existing "jsPlugins" registrations into the
// generated config, deduplicated and sorted.
func preserveJsPlugins(existing, generated *OxlintConfig) {
	unioned := slices.Clone(generated.JsPlugins)

	for _, pkg := range existing.JsPlugins {
		if !slices.Contains(unioned, pkg) {
			unioned = append(unioned, pkg)
		}
	}

	slices.Sort(unioned)
	generated.JsPlugins = unioned
}

// preserveExternalRules copies rules owned by known external plugins from
// the existing config into the generated one, verbatim including options.
func preserveExternalRules(existing, generated *OxlintConfig) {
	if generated.Rules == nil {
		generated.Rules = make(map[string]any)
	}

	for name, value := range existing.Rules {
		if _, external := rule.ExternalPluginByRuleName(name); external {
			generated.Rules[name] = value
		}
	}
}

// preserveExternalSettings copies each known external plugin's settings
// entry (e.g. "settings.shadcn") from the existing config into the
// generated one.
func preserveExternalSettings(existing, generated *OxlintConfig) {
	if len(existing.Settings) == 0 {
		return
	}

	if generated.Settings == nil {
		generated.Settings = make(map[string]any)
	}

	for _, external := range rule.KnownExternalPlugins() {
		if value, ok := existing.Settings[external.Prefix]; ok {
			generated.Settings[external.Prefix] = value
		}
	}
}

// HasExternalRules reports whether cfg contains any rules owned by a known
// external JS plugin.
func HasExternalRules(cfg *OxlintConfig) bool {
	for name := range cfg.Rules {
		if _, external := rule.ExternalPluginByRuleName(name); external {
			return true
		}
	}

	return false
}
