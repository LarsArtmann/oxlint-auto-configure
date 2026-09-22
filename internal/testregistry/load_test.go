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
	require.NotEmpty(t, reg.All())
	require.Equal(t, len(reg.All()), reg.Len())
}

func TestLoadIsRepeatable(t *testing.T) {
	t.Parallel()

	first := Load(t)
	second := Load(t)

	require.Len(t, first.All(), len(second.All()))
}
