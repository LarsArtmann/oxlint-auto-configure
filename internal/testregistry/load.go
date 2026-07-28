package testregistry

import (
	"testing"

	"github.com/larsartmann/oxlint-auto-configure/pkg/rule"
	"github.com/stretchr/testify/require"
)

// Load returns the rule.Registry used by tests across packages.
// It fails the test immediately if the embedded rules data cannot be decoded.
func Load(t *testing.T) *rule.Registry {
	t.Helper()

	reg, err := rule.LoadRegistry()
	require.NoError(t, err)

	return reg
}
