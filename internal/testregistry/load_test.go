package testregistry

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadReturnsNonEmptyRegistry(t *testing.T) {
	t.Parallel()

	reg := Load(t)
	require.NotNil(t, reg)
	require.NotEmpty(t, reg.All())
}

func TestLoadReturnsAllEmbeddedRules(t *testing.T) {
	t.Parallel()

	reg := Load(t)

	version, err := embeddedVersion()
	require.NoError(t, err)

	require.Equal(t, version.Total, len(reg.All()))
}

func TestLoadIsRepeatable(t *testing.T) {
	t.Parallel()

	first := Load(t)
	second := Load(t)

	require.Equal(t, len(first.All()), len(second.All()))
}
