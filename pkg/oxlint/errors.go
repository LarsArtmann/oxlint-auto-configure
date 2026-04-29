// Package oxlint provides integration with the oxlint linter binary.
package oxlint

import "errors"

// Sentinel errors for programmatic error handling.
var (
	// ErrNotFound is returned when the oxlint binary is not in PATH.
	ErrNotFound = errors.New("oxlint not found")
)
