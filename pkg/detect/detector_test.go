package detect

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/larsartmann/oxlint-auto-configure/pkg/profile"
	"github.com/larsartmann/oxlint-auto-configure/pkg/rule"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectUnknownProject(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	det := NewDetector(dir)
	pc, types, err := det.Detect()
	require.NoError(t, err)
	assert.Equal(t, profile.PluginConfig{}, pc)
	assert.NotEmpty(t, types)
}

func TestDetectReactProject(t *testing.T) {
	t.Parallel()
	pc, types := detectWithPackageJSON(t, `{"dependencies": {"react": "^18.0.0", "react-dom": "^18.0.0"}}`)

	assert.True(t, pc[rule.PluginReact])
	assert.True(t, pc[rule.PluginJSXA11y])
	assert.True(t, pc[rule.PluginReactPerf])

	assert.True(t, containsType(types, ProjectTypeReact))
}

func TestDetectNextJSProject(t *testing.T) {
	t.Parallel()
	pc, types := detectWithPackageJSON(t, `{"dependencies": {"next": "^14.0.0", "react": "^18.0.0"}}`)

	assert.True(t, pc[rule.PluginNextJS])
	assert.True(t, pc[rule.PluginReact])
	assert.True(t, containsType(types, ProjectTypeNextJS))
}

func TestDetectVueProject(t *testing.T) {
	t.Parallel()
	pc, types := detectWithPackageJSON(t, `{"dependencies": {"vue": "^3.0.0"}}`)

	assert.True(t, pc[rule.PluginVue])
	assert.True(t, containsType(types, ProjectTypeVue))
}

func TestDetectJestProject(t *testing.T) {
	t.Parallel()
	pc, types := detectWithPackageJSON(t, `{"devDependencies": {"jest": "^29.0.0"}}`)

	assert.True(t, pc[rule.PluginJest])
	assert.True(t, pc[rule.PluginNode])
	assert.True(t, containsType(types, ProjectTypeTest), "expected ProjectTypeTest in detected types")
}

func TestDetectVitestProject(t *testing.T) {
	t.Parallel()
	pc, types := detectWithPackageJSON(t, `{"devDependencies": {"vitest": "^1.0.0"}}`)

	assert.True(t, pc[rule.PluginVitest])
	assert.True(t, pc[rule.PluginNode])
	assert.True(t, containsType(types, ProjectTypeTest), "expected ProjectTypeTest in detected types")
}

func TestDetectTSProject(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "tsconfig.json"), []byte("{}"), 0o644)
	require.NoError(t, err)

	det := NewDetector(dir)
	_, types, err := det.Detect()
	require.NoError(t, err)

	var hasTS bool
	for _, t := range types {
		if t == ProjectTypePlainTS {
			hasTS = true
		}
	}
	assert.True(t, hasTS)
}

func TestFormatTypes(t *testing.T) {
	t.Parallel()
	assert.Equal(
		t,
		"react, nextjs",
		FormatTypes([]ProjectType{ProjectTypeReact, ProjectTypeNextJS}),
	)
	assert.Equal(t, "unknown", FormatTypes(nil))
}

func TestDetectPromiseProject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		pkgJSON string
	}{
		{"bluebird", `{"dependencies": {"bluebird": "^3.0.0"}}`},
		{"es6-promise", `{"dependencies": {"es6-promise": "^4.0.0"}}`},
		{"q", `{"dependencies": {"q": "^1.0.0"}}`},
		{"rsvp", `{"dependencies": {"rsvp": "^4.0.0"}}`},
		{"promise in name", `{"dependencies": {"promise-polyfill": "^1.0.0"}}`},
		{"core-js", `{"dependencies": {"core-js": "^3.0.0"}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pc, _ := detectWithPackageJSON(t, tt.pkgJSON)
			assert.True(t, pc[rule.PluginPromise], "expected promise plugin for %s", tt.name)
		})
	}
}

func TestDetectNoFalsePromise(t *testing.T) {
	t.Parallel()
	pc, _ := detectWithPackageJSON(t, `{"dependencies": {"lodash": "^4.0.0"}}`)
	assert.False(t, pc[rule.PluginPromise], "should not detect promise plugin from lodash")
}

func TestProjectTypeString(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "react", string(ProjectTypeReact))
	assert.Equal(t, "nextjs", string(ProjectTypeNextJS))
	assert.Equal(t, "vue", string(ProjectTypeVue))
}

// detectWithPackageJSON writes pkgJSON to a fresh temp dir's package.json and
// returns the Detector's results. Fails the test on I/O or Detect errors.
func detectWithPackageJSON(t *testing.T, pkgJSON string) (profile.PluginConfig, []ProjectType) {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0o644))

	pc, types, err := NewDetector(dir).Detect()
	require.NoError(t, err)
	return pc, types
}

// containsType reports whether want appears in types.
func containsType(types []ProjectType, want ProjectType) bool {
	return slices.Contains(types, want)
}
