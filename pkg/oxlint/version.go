package oxlint

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var versionRegex = regexp.MustCompile(`(\d+\.\d+\.\d+)`)

// MinVersionForJsPlugins is the first oxlint version that supports loading
// external JS plugins via the "jsPlugins" config key (per @shadcn/lint's
// documented requirement).
const MinVersionForJsPlugins = "1.80.0"

// VersionAtLeast reports whether version (a plain "major.minor.patch"
// string) is greater than or equal to min. Non-numeric parts beyond the
// first three are ignored; missing parts count as zero. Malformed versions
// compare as zero, so callers can pass user-visible strings through.
func VersionAtLeast(version, min string) bool {
	return compareSemver(version, min) >= 0
}

// compareSemver returns -1, 0, or 1 comparing two numeric semver strings.
func compareSemver(a, b string) int {
	aParts, bParts := parseSemver(a), parseSemver(b)

	for i := range max(len(aParts), len(bParts)) {
		av, bv := semverPart(aParts, i), semverPart(bParts, i)

		switch {
		case av < bv:
			return -1
		case av > bv:
			return 1
		}
	}

	return 0
}

func parseSemver(v string) []string {
	// Keep at most three components; strip any non-numeric suffix (e.g.
	// prerelease or build metadata) hanging off a component.
	rest := strings.SplitN(v, ".", 3)
	for i, part := range rest {
		if nonDigit := strings.IndexFunc(part, func(r rune) bool {
			return r < '0' || r > '9'
		}); nonDigit >= 0 {
			rest[i] = part[:nonDigit]
		}
	}

	return rest
}

func semverPart(parts []string, i int) int {
	if i >= len(parts) || parts[i] == "" {
		return 0
	}

	n, err := strconv.Atoi(parts[i])
	if err != nil {
		return 0
	}

	return n
}

// CheckVersion runs `oxlint --version` and returns the version string.
func CheckVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "oxlint", "--version")

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrNotFound, err)
	}

	version := strings.TrimSpace(string(output))

	match := versionRegex.FindString(version)
	if match == "" {
		return "", fmt.Errorf("%w: %s", ErrUnexpectedVersionOutput, version)
	}

	return match, nil
}

// CheckBinary verifies oxlint is available in PATH.
func CheckBinary(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "oxlint", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w in PATH: %w", ErrNotFound, err)
	}

	return nil
}
