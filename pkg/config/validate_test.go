package config

import (
	"testing"

	"github.com/larsartmann/oxlint-auto-configure/internal/testregistry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testRuleNoDebugger = "no-debugger"

func TestValidateConfigValid(t *testing.T) {
	t.Parallel()
	reg := testregistry.Load(t)

	cfg := &OxlintConfig{
		Rules: map[string]any{testRuleNoDebugger: SeverityError},
	}

	result, err := ValidateConfig(cfg, reg)
	require.NoError(t, err)
	assert.Empty(t, result.UnknownRules)
	assert.Empty(t, result.InvalidSeverities)
	assert.Equal(t, 1, result.EnabledCount)
	assert.Equal(t, 0, result.DisabledCount)
}

func TestValidateConfigUnknownRules(t *testing.T) {
	t.Parallel()
	reg := testregistry.Load(t)

	cfg := &OxlintConfig{
		Rules: map[string]any{"nonexistent-rule-xyz": SeverityError},
	}

	result, err := ValidateConfig(cfg, reg)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unknown rules")
}

func TestValidateConfigInvalidSeverity(t *testing.T) {
	t.Parallel()
	reg := testregistry.Load(t)

	cfg := &OxlintConfig{
		Rules: map[string]any{testRuleNoDebugger: "badsev"},
	}

	result, err := ValidateConfig(cfg, reg)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid severities")
}

func TestValidateConfigDisabledRule(t *testing.T) {
	t.Parallel()
	reg := testregistry.Load(t)

	cfg := &OxlintConfig{
		Rules: map[string]any{testRuleNoDebugger: SeverityOff},
	}

	result, err := ValidateConfig(cfg, reg)
	require.NoError(t, err)
	assert.Equal(t, 0, result.EnabledCount)
	assert.Equal(t, 1, result.DisabledCount)
}

func TestValidateConfigEmptyRules(t *testing.T) {
	t.Parallel()
	reg := testregistry.Load(t)

	cfg := &OxlintConfig{Rules: map[string]any{}}

	result, err := ValidateConfig(cfg, reg)
	require.NoError(t, err)
	assert.Equal(t, 0, result.EnabledCount)
	assert.Equal(t, 0, result.DisabledCount)
}

func TestValidateConfigErrInvalidConfigWrapped(t *testing.T) {
	t.Parallel()
	reg := testregistry.Load(t)

	cfg := &OxlintConfig{
		Rules: map[string]any{"nonexistent-rule-xyz": SeverityError},
	}

	_, err := ValidateConfig(cfg, reg)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidConfig)
}

func TestValidateConfigExternalRulesExemptFromUnknownCheck(t *testing.T) {
	t.Parallel()
	reg := testregistry.Load(t)

	cfg := &OxlintConfig{
		Rules: map[string]any{
			"shadcn/no-restyle":          []any{"error", map[string]any{"allow": []any{"layout"}}},
			"shadcn/no-arbitrary-values": "error",
		},
	}

	result, err := ValidateConfig(cfg, reg)
	require.NoError(t, err,
		"rules owned by external JS plugins are valid even though the embedded registry cannot know them")
	assert.Empty(t, result.UnknownRules)
	assert.Equal(t,
		[]string{"shadcn/no-arbitrary-values", "shadcn/no-restyle"},
		result.ExternalRules,
	)
	assert.Equal(t, 2, result.EnabledCount)
}

func TestValidateConfigArraySeverityForms(t *testing.T) {
	t.Parallel()
	reg := testregistry.Load(t)

	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{"severity with options", []any{"error", map[string]any{"allow": []any{"layout"}}}, false},
		{"off with options", []any{"off", map[string]any{}}, false},
		{"empty array", []any{}, true},
		{"non-string first element", []any{42}, true},
		{"invalid severity with options", []any{"bogus", map[string]any{}}, true},
		{"bare number", 42, true},
		{"null", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &OxlintConfig{Rules: map[string]any{"shadcn/no-raw-colors": tt.value}}

			_, err := ValidateConfig(cfg, reg)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
