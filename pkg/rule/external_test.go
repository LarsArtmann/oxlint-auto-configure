package rule

import (
	"slices"
	"testing"
)

func TestKnownExternalPluginsIncludesShadcn(t *testing.T) {
	t.Parallel()

	plugins := KnownExternalPlugins()

	found := slices.ContainsFunc(plugins, func(p ExternalPlugin) bool {
		return p.Package == "@shadcn/lint" && p.Prefix == "shadcn"
	})
	if !found {
		t.Errorf("KnownExternalPlugins() = %v, want @shadcn/lint (shadcn) entry", plugins)
	}
}

func TestKnownExternalPluginsCloneIsIndependent(t *testing.T) {
	t.Parallel()

	plugins := KnownExternalPlugins()
	plugins[0].Package = "mutated"

	if again := KnownExternalPlugins(); again[0].Package == "mutated" {
		t.Error("KnownExternalPlugins() returned a slice that callers can mutate")
	}
}

func TestExternalPluginByPackage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		pkg       string
		want      ExternalPlugin
		wantFound bool
	}{
		{"@shadcn/lint", ExternalPlugin{Package: "@shadcn/lint", Prefix: "shadcn"}, true},
		{"shadcn", ExternalPlugin{}, false},
		{"", ExternalPlugin{}, false},
	}

	for _, tt := range tests {
		got, found := ExternalPluginByPackage(tt.pkg)
		if got != tt.want || found != tt.wantFound {
			t.Errorf("ExternalPluginByPackage(%q) = %v, %v; want %v, %v",
				tt.pkg, got, found, tt.want, tt.wantFound)
		}
	}
}

func TestExternalPluginByRuleName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		wantFound bool
	}{
		{"shadcn/no-restyle", true},
		{"shadcn/no-arbitrary-values", true},
		{"shadcn/", false}, // empty rule name after prefix
		{"shadcn", false},  // bare prefix, not a rule reference
		{"eslint/no-console", false},
		{"no-console", false}, // unprefixed rules never match
		{"", false},
	}

	for _, tt := range tests {
		_, found := ExternalPluginByRuleName(tt.name)
		if found != tt.wantFound {
			t.Errorf("ExternalPluginByRuleName(%q) found = %v, want %v", tt.name, found, tt.wantFound)
		}
	}
}
