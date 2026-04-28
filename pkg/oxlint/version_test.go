package oxlint

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckVersion(t *testing.T) {
	t.Parallel()

	ver, err := CheckVersion(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, ver)
	// Should be a semver like "1.59.0"
	assert.Regexp(t, `^\d+\.\d+\.\d+$`, ver)
}

func TestCheckBinary(t *testing.T) {
	t.Parallel()

	err := CheckBinary(context.Background())
	require.NoError(t, err)
}
