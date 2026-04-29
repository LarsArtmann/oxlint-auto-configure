// Package oxlint provides integration with the oxlint linter binary.
package oxlint

import "errors"

// Sentinel errors for programmatic error handling.
var (
	// ErrNotFound is returned when the oxlint binary is not in PATH.
	ErrNotFound = errors.New("oxlint not found")

	// ErrInvalidProfile is returned when an unknown profile name is given.
	ErrInvalidProfile = errors.New("invalid profile")

	// ErrInvalidConfig is returned when a config file cannot be parsed.
	ErrInvalidConfig = errors.New("invalid config")
)
