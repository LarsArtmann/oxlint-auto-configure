package oxlint

import "errors"

// Sentinel errors for programmatic error handling.
var (
	// ErrNotFound is returned when the oxlint binary is not in PATH.
	ErrNotFound = errors.New("oxlint not found")

	// ErrUnexpectedVersionOutput is returned when `oxlint --version`
	// produces output that does not match the expected version format.
	ErrUnexpectedVersionOutput = errors.New("unexpected oxlint version output")

	// ErrOxlintStderr is returned when oxlint exits with a non-zero status
	// and writes to stderr.
	ErrOxlintStderr = errors.New("oxlint stderr output")
)
