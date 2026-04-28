package detect

import (
	"os"
	"path/filepath"
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
	dir := t.TempDir()
	pkgJSON := `{"dependencies": {"react": "^18.0.0", "react-dom": "^18.0.0"}}`
	err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0o644)
	require.NoError(t, err)

	det := NewDetector(dir)
	pc, types, err := det.Detect()
	require.NoError(t, err)

	assert.True(t, pc[rule.PluginReact])
	assert.True(t, pc[rule.PluginJSXA11y])
	assert.True(t, pc[rule.PluginReactPerf])

	var hasReact bool
	for _, t := range types {
		if t == ProjectTypeReact {
			hasReact = true
		}
	}
	assert.True(t, hasReact)
}

func TestDetectNextJSProject(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	pkgJSON := `{"dependencies": {"next": "^14.0.0", "react": "^18.0.0"}}`
	err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0o644)
	require.NoError(t, err)

	det := NewDetector(dir)
	pc, types, err := det.Detect()
	require.NoError(t, err)

	assert.True(t, pc[rule.PluginNextJS])
	assert.True(t, pc[rule.PluginReact])

	var hasNextJS bool
	for _, t := range types {
		if t == ProjectTypeNextJS {
			hasNextJS = true
		}
	}
	assert.True(t, hasNextJS)
}

func TestDetectVueProject(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	pkgJSON := `{"dependencies": {"vue": "^3.0.0"}}`
	err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0o644)
	require.NoError(t, err)

	det := NewDetector(dir)
	pc, types, err := det.Detect()
	require.NoError(t, err)

	assert.True(t, pc[rule.PluginVue])

	var hasVue bool
	for _, t := range types {
		if t == ProjectTypeVue {
			hasVue = true
		}
	}
	assert.True(t, hasVue)
}

func TestDetectJestProject(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	pkgJSON := `{"devDependencies": {"jest": "^29.0.0"}}`
	err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0o644)
	require.NoError(t, err)

	det := NewDetector(dir)
	pc, _, err := det.Detect()
	require.NoError(t, err)

	assert.True(t, pc[rule.PluginJest])
	assert.True(t, pc[rule.PluginNode])
}

func TestDetectVitestProject(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	pkgJSON := `{"devDependencies": {"vitest": "^1.0.0"}}`
	err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0o644)
	require.NoError(t, err)

	det := NewDetector(dir)
	pc, _, err := det.Detect()
	require.NoError(t, err)

	assert.True(t, pc[rule.PluginVitest])
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

func TestProjectTypeString(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "react", string(ProjectTypeReact))
	assert.Equal(t, "nextjs", string(ProjectTypeNextJS))
	assert.Equal(t, "vue", string(ProjectTypeVue))
}
