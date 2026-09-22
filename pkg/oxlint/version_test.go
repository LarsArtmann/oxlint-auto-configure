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

func TestVersionAtLeast(t *testing.T) {
	t.Parallel()

	tests := []struct {
		version string
		min     string
		want    bool
	}{
		{"1.80.0", "1.80.0", true},
		{"1.80.1", "1.80.0", true},
		{"1.81.0", "1.80.0", true},
		{"2.0.0", "1.80.0", true},
		{"1.79.9", "1.80.0", false},
		{"1.73.0", "1.80.0", false},
		{"1.8.0", "1.80.0", false}, // digit-wise, not lexicographic
		{"1.80", "1.80.0", true},   // missing patch counts as zero
		{"1.80.0-beta.1", "1.80.0", true},
		{"", "1.80.0", false},
		{"garbage", "1.80.0", false},
	}

	for _, tt := range tests {
		if got := VersionAtLeast(tt.version, tt.min); got != tt.want {
			t.Errorf("VersionAtLeast(%q, %q) = %v, want %v", tt.version, tt.min, got, tt.want)
		}
	}
}
