// Package config generates .oxlintrc.json configuration files.
package config

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"

	"github.com/larsartmann/oxlint-auto-configure/pkg/profile"
	"github.com/larsartmann/oxlint-auto-configure/pkg/rule"
)

// OxlintConfig represents the .oxlintrc.json structure.
type OxlintConfig struct {
	Plugins    []string         `json:"plugins,omitempty"`
	Categories map[string]string `json:"categories,omitempty"`
	Rules      map[string]string `json:"rules,omitempty"`
	Settings   map[string]any   `json:"settings,omitempty"`
	Env        map[string]bool  `json:"env,omitempty"`
}

// Generator creates oxlint configuration from profile decisions.
type Generator struct {
	categorizer *profile.Categorizer
	registry    *rule.Registry
}

// NewGenerator creates a config generator.
func NewGenerator(c *profile.Categorizer, r *rule.Registry) *Generator {
	return &Generator{categorizer: c, registry: r}
}

// Generate creates an OxlintConfig based on the categorizer's decisions.
func (g *Generator) Generate() *OxlintConfig {
	cfg := &OxlintConfig{
		Plugins:    g.enabledPlugins(),
		Categories: g.categorySeverityMap(),
		Rules:      g.ruleSeverityMap(),
		Settings:   defaultSettings(),
		Env:        map[string]bool{"builtin": true},
	}
	return cfg
}

// GenerateAllError creates a config where ALL rules are set to "error".
// This is the maximal-typesafe profile.
func (g *Generator) GenerateAllError() *OxlintConfig {
	cfg := &OxlintConfig{
		Plugins: g.allPlugins(),
		Categories: map[string]string{
			"correctness": "error",
			"suspicious":  "error",
			"pedantic":    "error",
			"perf":        "error",
			"style":       "error",
			"restriction": "error",
			"nursery":     "warn",
		},
		Rules:    map[string]string{},
		Settings: defaultSettings(),
		Env:      map[string]bool{"builtin": true},
	}
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

// enabledPlugins returns the list of plugins that should be enabled.
func (g *Generator) enabledPlugins() []string {
	var plugins []string

	plugins = append(plugins, "unicorn", "typescript", "oxc")

	if g.categorizer != nil {
		pc := g.categorizer.EnabledPlugins()
		for _, p := range pc {
			name := pluginFlagToName(p)
			if name != "" && !slices.Contains(plugins, name) {
				plugins = append(plugins, name)
			}
		}
	}

	sort.Strings(plugins)
	return plugins
}

// allPlugins returns all possible plugins.
func (g *Generator) allPlugins() []string {
	plugins := []string{
		"eslint", "import", "jest", "jsdoc", "jsx_a11y",
		"nextjs", "node", "oxc", "promise", "react",
		"react_perf", "typescript", "unicorn", "vitest", "vue",
	}
	sort.Strings(plugins)
	return plugins
}

// categorySeverityMap returns the category→severity mapping for the profile.
func (g *Generator) categorySeverityMap() map[string]string {
	if g.categorizer == nil {
		return map[string]string{"correctness": "error"}
	}

	cats := map[string]rule.Category{
		"correctness": rule.CategoryCorrectness,
		"suspicious":  rule.CategorySuspicious,
		"pedantic":    rule.CategoryPedantic,
		"perf":        rule.CategoryPerf,
		"style":       rule.CategoryStyle,
		"restriction": rule.CategoryRestriction,
		"nursery":     rule.CategoryNursery,
	}

	result := make(map[string]string)
	for name, cat := range cats {
		rules := g.registry.ByCategory(cat)
		if len(rules) > 0 {
			decision := g.categorizer.Decide(rules[0])
			result[name] = string(decision)
		}
	}
	return result
}

// ruleSeverityMap returns per-rule overrides that differ from the category default.
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

		fullName := d.Rule.FullName()
		catSev, hasCat := catMap[string(d.Rule.Category)]

		if d.Severity == rule.SeverityOff {
			rules[fullName] = "off"
		} else if hasCat && string(d.Severity) != catSev {
			rules[fullName] = string(d.Severity)
		}
	}

	return rules
}

func defaultSettings() map[string]any {
	return map[string]any{
		"jsx-a11y": map[string]any{
			"polymorphicPropName": nil,
			"components":         map[string]any{},
			"attributes":         map[string]any{},
		},
		"next": map[string]any{
			"rootDir": []any{},
		},
		"react": map[string]any{
			"formComponents":            []any{},
			"linkComponents":            []any{},
			"version":                   nil,
			"componentWrapperFunctions": []any{},
		},
		"jsdoc": map[string]any{
			"ignorePrivate":                    false,
			"ignoreInternal":                   false,
			"ignoreReplacesDocs":               true,
			"overrideReplacesDocs":             true,
			"augmentsExtendsReplacesDocs":      false,
			"implementsReplacesDocs":           false,
			"exemptDestructuredRootsFromChecks": false,
			"tagNamePreference":               map[string]any{},
		},
		"vitest": map[string]any{
			"typecheck": false,
		},
	}
}

func pluginFlagToName(flag string) string {
	m := map[string]string{
		"--import-plugin":     "import",
		"--react-plugin":      "react",
		"--jsdoc-plugin":      "jsdoc",
		"--jest-plugin":       "jest",
		"--vitest-plugin":     "vitest",
		"--jsx-a11y-plugin":   "jsx_a11y",
		"--nextjs-plugin":     "nextjs",
		"--react-perf-plugin": "react_perf",
		"--promise-plugin":    "promise",
		"--node-plugin":       "node",
		"--vue-plugin":        "vue",
	}
	return m[flag]
}
