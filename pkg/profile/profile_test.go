package profile

import (
	"testing"

	"github.com/larsartmann/oxlint-auto-configure/pkg/rule"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadTestRegistry(t *testing.T) *rule.Registry {
	t.Helper()

	reg, err := rule.LoadRegistry()
	require.NoError(t, err)

	return reg
}

func TestProfileString(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "maximal-typesafe", ProfileMaximalTypesafe.String())
	assert.Equal(t, "recommended", ProfileRecommended.String())
	assert.Equal(t, "strict", ProfileStrict.String())
	assert.Equal(t, "minimal", ProfileMinimal.String())
}

func TestProfileIsValid(t *testing.T) {
	t.Parallel()
	assert.True(t, ProfileMaximalTypesafe.IsValid())
	assert.True(t, ProfileRecommended.IsValid())
	assert.True(t, ProfileStrict.IsValid())
	assert.True(t, ProfileMinimal.IsValid())
	assert.False(t, Profile("invalid").IsValid())
}

func TestProfileDescription(t *testing.T) {
	t.Parallel()
	assert.NotEmpty(t, ProfileMaximalTypesafe.Description())
	assert.NotEmpty(t, ProfileRecommended.Description())
	assert.NotEmpty(t, ProfileStrict.Description())
	assert.NotEmpty(t, ProfileMinimal.Description())
	assert.Equal(t, "unknown profile", Profile("invalid").Description())
}

func TestCategorizerMaximalTypesafe(t *testing.T) {
	t.Parallel()

	cat := NewCategorizer(ProfileMaximalTypesafe, PluginConfig{})

	assert.Equal(t, rule.SeverityError, cat.Decide(rule.Rule{Category: rule.CategoryCorrectness}))
	assert.Equal(t, rule.SeverityError, cat.Decide(rule.Rule{Category: rule.CategoryStyle}))
	assert.Equal(t, rule.SeverityError, cat.Decide(rule.Rule{Category: rule.CategoryPerf}))
	assert.Equal(t, rule.SeverityWarn, cat.Decide(rule.Rule{Category: rule.CategoryNursery}))
	assert.Equal(t, rule.SeverityError, cat.Decide(rule.Rule{Category: rule.CategoryRestriction}))
}

func TestCategorizerRecommended(t *testing.T) {
	t.Parallel()

	cat := NewCategorizer(ProfileRecommended, PluginConfig{})

	assert.Equal(t, rule.SeverityError, cat.Decide(rule.Rule{Category: rule.CategoryCorrectness}))
	assert.Equal(t, rule.SeverityError, cat.Decide(rule.Rule{Category: rule.CategorySuspicious}))
	assert.Equal(t, rule.SeverityWarn, cat.Decide(rule.Rule{Category: rule.CategoryStyle}))
	assert.Equal(t, rule.SeverityWarn, cat.Decide(rule.Rule{Category: rule.CategoryPerf}))
	assert.Equal(t, rule.SeverityWarn, cat.Decide(rule.Rule{Category: rule.CategoryPedantic}))
	assert.Equal(t, rule.SeverityWarn, cat.Decide(rule.Rule{Category: rule.CategoryRestriction}))
	assert.Equal(t, rule.SeverityOff, cat.Decide(rule.Rule{Category: rule.CategoryNursery}))
}

func TestCategorizerStrict(t *testing.T) {
	t.Parallel()

	cat := NewCategorizer(ProfileStrict, PluginConfig{})

	assert.Equal(t, rule.SeverityError, cat.Decide(rule.Rule{Category: rule.CategoryCorrectness}))
	assert.Equal(t, rule.SeverityError, cat.Decide(rule.Rule{Category: rule.CategorySuspicious}))
	assert.Equal(t, rule.SeverityWarn, cat.Decide(rule.Rule{Category: rule.CategoryStyle}))
	assert.Equal(t, rule.SeverityWarn, cat.Decide(rule.Rule{Category: rule.CategoryPedantic}))
	assert.Equal(t, rule.SeverityOff, cat.Decide(rule.Rule{Category: rule.CategoryNursery}))
}

func TestDecideCategory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		profile  Profile
		category rule.Category
		want     rule.SeverityDecision
		include  bool
	}{
		{
			"maximal correctness",
			ProfileMaximalTypesafe,
			rule.CategoryCorrectness,
			rule.SeverityError,
			true,
		},
		{"maximal nursery", ProfileMaximalTypesafe, rule.CategoryNursery, rule.SeverityWarn, true},
		{
			"recommended correctness",
			ProfileRecommended,
			rule.CategoryCorrectness,
			rule.SeverityError,
			true,
		},
		{"recommended style", ProfileRecommended, rule.CategoryStyle, rule.SeverityWarn, true},
		{"recommended nursery", ProfileRecommended, rule.CategoryNursery, rule.SeverityOff, true},
		{"strict correctness", ProfileStrict, rule.CategoryCorrectness, rule.SeverityError, true},
		{"strict style", ProfileStrict, rule.CategoryStyle, rule.SeverityWarn, true},
		{"strict nursery", ProfileStrict, rule.CategoryNursery, rule.SeverityOff, true},
		{"minimal correctness", ProfileMinimal, rule.CategoryCorrectness, rule.SeverityError, true},
		{"minimal style omitted", ProfileMinimal, rule.CategoryStyle, rule.SeverityOff, false},
		{"minimal nursery omitted", ProfileMinimal, rule.CategoryNursery, rule.SeverityOff, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cat := NewCategorizer(tt.profile, PluginConfig{})
			got, include := cat.DecideCategory(tt.category)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.include, include)
		})
	}
}

func TestCategorizerMinimal(t *testing.T) {
	t.Parallel()

	cat := NewCategorizer(ProfileMinimal, PluginConfig{})

	assert.Equal(
		t,
		rule.SeverityError,
		cat.Decide(rule.Rule{Category: rule.CategoryCorrectness, Enabled: true}),
	)
	assert.Equal(
		t,
		rule.SeverityWarn,
		cat.Decide(rule.Rule{Category: rule.CategoryStyle, Enabled: true}),
	)
	assert.Equal(
		t,
		rule.SeverityOff,
		cat.Decide(rule.Rule{Category: rule.CategoryStyle, Enabled: false}),
	)
}

func TestPluginRelevance(t *testing.T) {
	t.Parallel()

	cat := NewCategorizer(ProfileRecommended, PluginConfig{
		rule.PluginReact:  true,
		rule.PluginNextJS: true,
		rule.PluginJest:   true,
		rule.PluginVitest: true,
	})

	assert.True(t, cat.IsPluginRelevant(rule.Rule{Plugin: rule.PluginESLint}))
	assert.True(t, cat.IsPluginRelevant(rule.Rule{Plugin: rule.PluginTypeScript}))
	assert.True(t, cat.IsPluginRelevant(rule.Rule{Plugin: rule.PluginUnicorn}))
	assert.True(t, cat.IsPluginRelevant(rule.Rule{Plugin: rule.PluginOXC}))
	assert.True(t, cat.IsPluginRelevant(rule.Rule{Plugin: rule.PluginReact}))
	assert.True(t, cat.IsPluginRelevant(rule.Rule{Plugin: rule.PluginNextJS}))
	assert.True(t, cat.IsPluginRelevant(rule.Rule{Plugin: rule.PluginJest}))
	assert.True(t, cat.IsPluginRelevant(rule.Rule{Plugin: rule.PluginVitest}))
	assert.False(t, cat.IsPluginRelevant(rule.Rule{Plugin: rule.PluginVue}))
	assert.False(t, cat.IsPluginRelevant(rule.Rule{Plugin: rule.PluginJSDoc}))
}

func TestEnabledPlugins(t *testing.T) {
	t.Parallel()

	cat := NewCategorizer(ProfileRecommended, PluginConfig{
		rule.PluginReact:   true,
		rule.PluginJest:    true,
		rule.PluginJSXA11y: true,
	})

	plugins := cat.EnabledPlugins()
	assert.Contains(t, plugins, rule.PluginReact)
	assert.Contains(t, plugins, rule.PluginJest)
	assert.Contains(t, plugins, rule.PluginJSXA11y)
}

func TestDecideAll(t *testing.T) {
	t.Parallel()
	reg := loadTestRegistry(t)

	cat := NewCategorizer(ProfileRecommended, PluginConfig{})
	decisions := cat.DecideAll(reg)

	assert.NotEmpty(t, decisions)
	assert.Len(t, decisions, reg.Len())
}

func TestAllProfiles(t *testing.T) {
	t.Parallel()

	profiles := AllProfiles()
	assert.Len(t, profiles, 4)
}

func TestDecideCategoryUnknown(t *testing.T) {
	t.Parallel()

	unknownCat := rule.Category("unknown_category")

	t.Run("minimal", func(t *testing.T) {
		t.Parallel()

		c := NewCategorizer(ProfileMinimal, nil)
		_, hasCat := c.DecideCategory(unknownCat)
		assert.False(t, hasCat, "unknown category should be omitted for minimal")
	})

	t.Run("maximal-typesafe", func(t *testing.T) {
		t.Parallel()

		c := NewCategorizer(ProfileMaximalTypesafe, nil)
		_, hasCat := c.DecideCategory(unknownCat)
		assert.True(
			t,
			hasCat,
			"unknown category should be included for maximal-typesafe (default severity)",
		)
	})

	t.Run("recommended", func(t *testing.T) {
		t.Parallel()

		c := NewCategorizer(ProfileRecommended, nil)
		_, hasCat := c.DecideCategory(unknownCat)
		assert.True(t, hasCat, "unknown category gets default severity for recommended")
	})

	t.Run("strict", func(t *testing.T) {
		t.Parallel()

		c := NewCategorizer(ProfileStrict, nil)
		_, hasCat := c.DecideCategory(unknownCat)
		assert.True(t, hasCat, "unknown category gets default severity for strict")
	})
}
