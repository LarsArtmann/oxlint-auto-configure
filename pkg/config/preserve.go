package config

import (
	"encoding/json/v2"
	"fmt"
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
//
// "overrides" blocks are preserved wholesale for the same reason: the
// generator never emits them, so they are pure existing policy (e.g.
// @shadcn/lint's component-dir disables). Exact-duplicate blocks are
// collapsed; everything else survives verbatim, order preserved.
func PreserveExternal(existing, generated *OxlintConfig) *OxlintConfig {
	if existing == nil {
		return generated
	}

	preserveJsPlugins(existing, generated)
	preserveExternalRules(existing, generated)
	preserveExternalSettings(existing, generated)
	preserveOverrides(existing, generated)

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

// preserveOverrides merges the existing config's "overrides" blocks into
// the generated one, deduplicated by canonical JSON form: an exact duplicate
// (within either config) is kept once, first occurrence wins, order
// preserved. Blocks that cannot be marshaled cannot collide either, so they
// are kept verbatim rather than dropped.
func preserveOverrides(existing, generated *OxlintConfig) {
	if len(existing.Overrides) == 0 {
		return
	}

	merged := make([]map[string]any, 0, len(generated.Overrides)+len(existing.Overrides))
	seen := make(map[string]bool, cap(merged))

	for _, block := range append(slices.Clone(generated.Overrides), existing.Overrides...) {
		key, err := canonicalOverridesKey(block)
		if err != nil {
			merged = append(merged, block)

			continue
		}

		if seen[key] {
			continue
		}

		seen[key] = true
		merged = append(merged, block)
	}

	generated.Overrides = merged
}

// canonicalOverridesKey renders a single overrides block in a canonical form
// (sorted keys via JSON marshaling) so structurally identical blocks dedup
// regardless of key order.
func canonicalOverridesKey(block map[string]any) (string, error) {
	data, err := json.Marshal(block)
	if err != nil {
		return "", fmt.Errorf("canonicalize overrides block: %w", err)
	}

	return string(data), nil
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
