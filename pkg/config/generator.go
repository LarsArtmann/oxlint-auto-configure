// Package config generates .oxlintrc.json configuration files.
package config

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/larsartmann/oxlint-auto-configure/pkg/detect"
	"github.com/larsartmann/oxlint-auto-configure/pkg/profile"
	"github.com/larsartmann/oxlint-auto-configure/pkg/rule"
)

// OxlintConfig represents the .oxlintrc.json structure.
type OxlintConfig struct {
	Plugins    []string          `json:"plugins,omitempty"`
	Categories map[string]string `json:"categories,omitempty"`
	Rules      map[string]string `json:"rules,omitempty"`
	Settings   map[string]any    `json:"settings,omitempty"`
	Env        map[string]bool   `json:"env,omitempty"`
}

// Generator creates oxlint configuration from profile decisions.
type Generator struct {
	categorizer  *profile.Categorizer
	registry     *rule.Registry
	projectTypes []detect.ProjectType
}

// NewGenerator creates a config generator.
func NewGenerator(
	c *profile.Categorizer,
	r *rule.Registry,
	projectTypes []detect.ProjectType,
) *Generator {
	return &Generator{categorizer: c, registry: r, projectTypes: projectTypes}
}

// Generate creates an OxlintConfig based on the categorizer's decisions.
func (g *Generator) Generate() *OxlintConfig {
	cfg := &OxlintConfig{
		Plugins:    g.enabledPlugins(),
		Categories: g.categorySeverityMap(),
		Rules:      g.ruleSeverityMap(),
		Settings:   g.buildSettings(),
		Env:        g.buildEnv(),
	}
	return cfg
}

// GenerateMaximal creates a config for the maximal-typesafe profile (all plugins, all categories).
// This is the maximal-typesafe profile.
func (g *Generator) GenerateMaximal() *OxlintConfig {
	cfg := g.Generate()
	cfg.Plugins = g.allPlugins()
	return cfg
}

// ToJSON serializes the config to pretty-printed JSON.
func (c *OxlintConfig) ToJSON() ([]byte, error) {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal config: %w", err)
	}
	return data, nil
}

// FromJSON parses an oxlint config from JSON bytes.
func FromJSON(data []byte) (*OxlintConfig, error) {
	var cfg OxlintConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return &cfg, nil
}

// buildEnv returns the env map based on detected project types.
// Always includes "builtin"; adds "node" when a Node.js project is detected.
func (g *Generator) buildEnv() map[string]bool {
	env := map[string]bool{"builtin": true}
	for _, pt := range g.projectTypes {
		if pt == detect.ProjectTypeNode || pt == detect.ProjectTypeTest {
			env["node"] = true
		}
	}
	return env
}

// enabledPlugins returns the list of plugins that should be enabled.
func (g *Generator) enabledPlugins() []string {
	seen := make(map[string]bool)
	var plugins []string

	for _, p := range []rule.Plugin{rule.PluginOXC, rule.PluginTypeScript, rule.PluginUnicorn} {
		name := string(p)
		seen[name] = true
		plugins = append(plugins, name)
	}

	if g.categorizer != nil {
		for _, p := range g.categorizer.EnabledPlugins() {
			name := string(p)
			if !seen[name] {
				seen[name] = true
				plugins = append(plugins, name)
			}
		}
	}

	sort.Strings(plugins)
	return plugins
}

// allPlugins returns all possible plugins.
func (g *Generator) allPlugins() []string {
	plugins := make([]string, 0, len(rule.AllPlugins()))
	for _, p := range rule.AllPlugins() {
		plugins = append(plugins, string(p))
	}
	sort.Strings(plugins)
	return plugins
}

// categorySeverityMap returns the category→severity mapping for the profile.
// Uses Categorizer.DecideCategory instead of sampling rules, ensuring
// correct category severities even for profiles with mixed-severity categories.
func (g *Generator) categorySeverityMap() map[string]string {
	if g.categorizer == nil {
		return map[string]string{"correctness": "error"}
	}

	result := make(map[string]string)
	for _, cat := range rule.AllCategories() {
		severity, include := g.categorizer.DecideCategory(cat)
		if include {
			result[string(cat)] = string(severity)
		}
	}
	return result
}

// ruleSeverityMap returns per-rule overrides that differ from the category default.
// Only emits rules whose severity contradicts their category-level severity.
// Rules from omitted categories (e.g., minimal's non-correctness categories)
// are skipped entirely — oxlint defaults apply.
func (g *Generator) ruleSeverityMap() map[string]string {
	if g.categorizer == nil {
		return map[string]string{}
	}

	decisions := g.categorizer.DecideAll(g.registry)
	catMap := g.categorySeverityMap()
	rules := make(map[string]string)

	for _, d := range decisions {
		if !g.categorizer.IsPluginRelevant(d.Rule) {
			continue
		}

		catSev, hasCat := catMap[string(d.Rule.Category)]

		// Skip rules from omitted categories — oxlint defaults apply.
		if !hasCat {
			continue
		}

		// Only emit when the rule's severity differs from its category.
		if string(d.Severity) != catSev {
			rules[d.Rule.FullName()] = string(d.Severity)
		}
	}

	return rules
}

// pluginSettings maps plugins to their oxlint settings key and defaults.
var pluginSettings = map[rule.Plugin]struct { //nolint:gochecknoglobals // immutable lookup table
	key   string
	value map[string]any
}{
	rule.PluginJSXA11y: {"jsx-a11y", map[string]any{
		"polymorphicPropName": nil,
		"components":          map[string]any{},
		"attributes":          map[string]any{},
	}},
	rule.PluginNextJS: {"next", map[string]any{
		"rootDir": []any{},
	}},
	rule.PluginReact: {"react", map[string]any{
		"formComponents":            []any{},
		"linkComponents":            []any{},
		"version":                   nil,
		"componentWrapperFunctions": []any{},
	}},
	rule.PluginJSDoc: {"jsdoc", map[string]any{
		"ignorePrivate":                     false,
		"ignoreInternal":                    false,
		"ignoreReplacesDocs":                true,
		"overrideReplacesDocs":              true,
		"augmentsExtendsReplacesDocs":       false,
		"implementsReplacesDocs":            false,
		"exemptDestructuredRootsFromChecks": false,
		"tagNamePreference":                 map[string]any{},
	}},
	rule.PluginVitest: {"vitest", map[string]any{
		"typecheck": false,
	}},
}

// buildSettings returns settings only for enabled plugins.
func (g *Generator) buildSettings() map[string]any {
	settings := make(map[string]any)
	if g.categorizer == nil {
		return settings
	}

	for _, p := range g.categorizer.EnabledPlugins() {
		if ps, ok := pluginSettings[p]; ok {
			settings[ps.key] = ps.value
		}
	}
	return settings
}
