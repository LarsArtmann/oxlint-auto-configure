package config

import (
	"testing"

	"github.com/larsartmann/oxlint-auto-configure/pkg/rule"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateConfigValid(t *testing.T) {
	t.Parallel()
	reg, err := rule.LoadRegistry()
	require.NoError(t, err)

	cfg := &OxlintConfig{
		Rules: map[string]string{"no-debugger": "error"},
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
	reg, err := rule.LoadRegistry()
	require.NoError(t, err)

	cfg := &OxlintConfig{
		Rules: map[string]string{"nonexistent-rule-xyz": "error"},
	}

	result, err := ValidateConfig(cfg, reg)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unknown rules")
}

func TestValidateConfigInvalidSeverity(t *testing.T) {
	t.Parallel()
	reg, err := rule.LoadRegistry()
	require.NoError(t, err)

	cfg := &OxlintConfig{
		Rules: map[string]string{"no-debugger": "badsev"},
	}

	result, err := ValidateConfig(cfg, reg)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid severities")
}

func TestValidateConfigDisabledRule(t *testing.T) {
	t.Parallel()
	reg, err := rule.LoadRegistry()
	require.NoError(t, err)

	cfg := &OxlintConfig{
		Rules: map[string]string{"no-debugger": "off"},
	}

	result, err := ValidateConfig(cfg, reg)
	require.NoError(t, err)
	assert.Equal(t, 0, result.EnabledCount)
	assert.Equal(t, 1, result.DisabledCount)
}

func TestValidateConfigEmptyRules(t *testing.T) {
	t.Parallel()
	reg, err := rule.LoadRegistry()
	require.NoError(t, err)

	cfg := &OxlintConfig{Rules: map[string]string{}}

	result, err := ValidateConfig(cfg, reg)
	require.NoError(t, err)
	assert.Equal(t, 0, result.EnabledCount)
	assert.Equal(t, 0, result.DisabledCount)
}

func TestValidateConfigErrInvalidConfigWrapped(t *testing.T) {
	t.Parallel()
	reg, err := rule.LoadRegistry()
	require.NoError(t, err)

	cfg := &OxlintConfig{
		Rules: map[string]string{"nonexistent-rule-xyz": "error"},
	}

	_, err = ValidateConfig(cfg, reg)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidConfig)
}
