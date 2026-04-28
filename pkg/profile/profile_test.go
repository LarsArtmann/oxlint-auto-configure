package profile

import (
	"testing"

	"github.com/larsartmann/oxlint-auto-configure/pkg/rule"
	"github.com/stretchr/testify/assert"
)

func TestProfileIsValid(t *testing.T) {
	assert.True(t, ProfileMaximalTypesafe.IsValid())
	assert.True(t, ProfileRecommended.IsValid())
	assert.True(t, ProfileStrict.IsValid())
	assert.True(t, ProfileMinimal.IsValid())
	assert.False(t, Profile("invalid").IsValid())
}

func TestProfileDescription(t *testing.T) {
	assert.NotEmpty(t, ProfileMaximalTypesafe.Description())
	assert.NotEmpty(t, ProfileRecommended.Description())
	assert.NotEmpty(t, ProfileStrict.Description())
	assert.NotEmpty(t, ProfileMinimal.Description())
	assert.Equal(t, "unknown profile", Profile("invalid").Description())
}

func TestCategorizerMaximalTypesafe(t *testing.T) {
	cat := NewCategorizer(ProfileMaximalTypesafe, PluginConfig{})

	assert.Equal(t, rule.SeverityError, cat.Decide(rule.Rule{Category: rule.CategoryCorrectness}))
	assert.Equal(t, rule.SeverityError, cat.Decide(rule.Rule{Category: rule.CategoryStyle}))
	assert.Equal(t, rule.SeverityError, cat.Decide(rule.Rule{Category: rule.CategoryPerf}))
	assert.Equal(t, rule.SeverityWarn, cat.Decide(rule.Rule{Category: rule.CategoryNursery}))
	assert.Equal(t, rule.SeverityError, cat.Decide(rule.Rule{Category: rule.CategoryRestriction}))
}

func TestCategorizerRecommended(t *testing.T) {
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
	cat := NewCategorizer(ProfileStrict, PluginConfig{})

	assert.Equal(t, rule.SeverityError, cat.Decide(rule.Rule{Category: rule.CategoryCorrectness}))
	assert.Equal(t, rule.SeverityError, cat.Decide(rule.Rule{Category: rule.CategorySuspicious}))
	assert.Equal(t, rule.SeverityWarn, cat.Decide(rule.Rule{Category: rule.CategoryStyle}))
	assert.Equal(t, rule.SeverityWarn, cat.Decide(rule.Rule{Category: rule.CategoryPedantic}))
	assert.Equal(t, rule.SeverityOff, cat.Decide(rule.Rule{Category: rule.CategoryNursery}))
}

func TestCategorizerMinimal(t *testing.T) {
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
	cat := NewCategorizer(ProfileRecommended, PluginConfig{
		React:  true,
		NextJS: true,
		Jest:   true,
		Vitest: true,
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
	cat := NewCategorizer(ProfileRecommended, PluginConfig{
		React:   true,
		Jest:    true,
		JSXA11y: true,
	})

	plugins := cat.EnabledPlugins()
	assert.Contains(t, plugins, "--react-plugin")
	assert.Contains(t, plugins, "--jest-plugin")
	assert.Contains(t, plugins, "--jsx-a11y-plugin")
}

func TestDecideAll(t *testing.T) {
	reg, err := rule.LoadRegistry()
	assert.NoError(t, err)

	cat := NewCategorizer(ProfileRecommended, PluginConfig{})
	decisions := cat.DecideAll(reg)

	assert.NotEmpty(t, decisions)
	assert.Len(t, decisions, reg.Len())
}

func TestAllProfiles(t *testing.T) {
	profiles := AllProfiles()
	assert.Len(t, profiles, 4)
}
