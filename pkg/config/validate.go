package config

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/larsartmann/oxlint-auto-configure/pkg/rule"
)

// ErrInvalidConfig is returned when a config file cannot be validated.
var ErrInvalidConfig = errors.New("invalid config")

// ErrInvalidProfile is returned when an unknown profile name is given.
var ErrInvalidProfile = errors.New("invalid profile")

// ValidateResult contains the outcome of validating an OxlintConfig.
type ValidateResult struct {
	UnknownRules      []string
	InvalidSeverities []string
	EnabledCount      int
	DisabledCount     int
}

// ValidateConfig checks an OxlintConfig for unknown rules and invalid severities.
// Returns a ValidateResult with details, or an error if validation fails.
func ValidateConfig(cfg *OxlintConfig, reg *rule.Registry) (*ValidateResult, error) {
	result := &ValidateResult{
		UnknownRules:      nil,
		InvalidSeverities: nil,
		EnabledCount:      0,
		DisabledCount:     0,
	}

	if err := validateRules(cfg, reg, result); err != nil {
		return nil, err
	}

	if err := validateSeverities(cfg, result); err != nil {
		return nil, err
	}

	return result, nil
}

func validateRules(cfg *OxlintConfig, reg *rule.Registry, result *ValidateResult) error {
	for name := range cfg.Rules {
		if _, ok := reg.ByName(name); !ok {
			result.UnknownRules = append(result.UnknownRules, name)
		}
	}

	if len(result.UnknownRules) > 0 {
		slices.Sort(result.UnknownRules)

		return fmt.Errorf("%w: %d unknown rules (%s)",
			ErrInvalidConfig, len(result.UnknownRules),
			strings.Join(result.UnknownRules, ", "))
	}

	return nil
}

func validateSeverities(cfg *OxlintConfig, result *ValidateResult) error {
	for name, sev := range cfg.Rules {
		if !rule.SeverityDecision(sev).IsValid() {
			result.InvalidSeverities = append(result.InvalidSeverities,
				fmt.Sprintf("%s=%s", name, sev))
		}
	}

	if len(result.InvalidSeverities) > 0 {
		slices.Sort(result.InvalidSeverities)

		return fmt.Errorf("%w: %d invalid severities (%s)",
			ErrInvalidConfig, len(result.InvalidSeverities),
			strings.Join(result.InvalidSeverities, ", "))
	}

	for _, sev := range cfg.Rules {
		if sev != "off" {
			result.EnabledCount++
		}
	}

	result.DisabledCount = len(cfg.Rules) - result.EnabledCount

	return nil
}
