package provider_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/larsartmann/go-finding"
	toolsdk "github.com/larsartmann/go-finding/toolsdk"
	"github.com/larsartmann/oxlint-auto-configure/pkg/config"
	"github.com/larsartmann/oxlint-auto-configure/pkg/provider"
	"github.com/stretchr/testify/require"
)

// reactPackageJSON is a minimal package.json that the detector recognizes as
// a React project.
const reactPackageJSON = `{"name":"test-app","dependencies":{"react":"^18.0.0"}}`

// isolateRegistry snapshots the process-global toolsdk registry so a test
// cannot observe registrations from other packages' init functions.
func isolateRegistry(t *testing.T) {
	t.Helper()

	snapshot := toolsdk.SnapshotForTest()

	t.Cleanup(func() { toolsdk.RestoreForTest(snapshot) })
}

func TestProviderRegistered(t *testing.T) {
	t.Parallel()

	isolateRegistry(t)

	var specs []toolsdk.Spec

	for _, s := range toolsdk.All() {
		if s.Name == "oxlint-auto-configure" {
			specs = append(specs, s)
		}
	}

	require.Len(t, specs, 1, "provider must register exactly once")
	require.NotEmpty(t, specs[0].Description)
	require.NotNil(t, specs[0].Detect, "provider must detect")
	require.NotNil(t, specs[0].Repair, "provider must repair")
	require.Equal(t, []string{"oxlint"}, specs[0].DependsOn)
	require.Equal(t, "javascript", specs[0].Trigger.Language)
}

func TestDetect_MissingConfigWithKnownProjectType(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, dir, "package.json", reactPackageJSON)

	findings, err := provider.Provider.Detect.Detect(workingDirCtx(t, dir))

	require.NoError(t, err)
	require.Len(t, findings, 1)
	require.Equal(t, finding.RuleName("OXLOPT_CONFIG_MISSING"), findings[0].Rule)
	require.Equal(t, finding.ToolName("oxlint-auto-configure"), findings[0].ToolName)
	require.Equal(t, finding.SeverityWarning, findings[0].Severity)
	require.Equal(t, finding.FilePath(".oxlintrc.json"), findings[0].Position.File)
	require.Equal(t, finding.FixStrategyDirect, findings[0].FixStrategy)
}

func TestDetect_ExistingConfigNeverFlagged(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, dir, "package.json", reactPackageJSON)
	writeFile(t, dir, ".oxlintrc.json", `{"rules":{}}`)

	findings, err := provider.Provider.Detect.Detect(workingDirCtx(t, dir))

	require.NoError(t, err)
	require.Empty(t, findings, "an existing config must never be flagged for regeneration")
}

func TestDetect_BarePackageJSONIsNodeProject(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, dir, "package.json", `{"name":"unidentifiable"}`)

	findings, err := provider.Provider.Detect.Detect(workingDirCtx(t, dir))

	require.NoError(t, err)
	require.Len(t, findings, 1, "a package.json without framework deps is still a Node project")
}

func TestDetect_NoJSProjectMarkers(t *testing.T) {
	t.Parallel()

	findings, err := provider.Provider.Detect.Detect(workingDirCtx(t, t.TempDir()))

	require.NoError(t, err)
	require.Empty(t, findings, "a directory without any JS project marker must not be flagged")
}

func TestRepair_DryRunHoldsBackWrite(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, dir, "package.json", reactPackageJSON)
	ctx := toolsdk.WithDryRun(workingDirCtx(t, dir), true)

	result, err := provider.Provider.Repair.Repair(ctx)

	require.NoError(t, err)
	require.Contains(t, result.Description, "dry-run")
	require.NoFileExists(t, filepath.Join(dir, ".oxlintrc.json"), "dry-run must not write the config")
}

func TestRepair_WritesOptimalConfig(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, dir, "package.json", reactPackageJSON)

	result, err := provider.Provider.Repair.Repair(workingDirCtx(t, dir))

	require.NoError(t, err)
	require.Contains(t, result.Description, "wrote")

	data, err := os.ReadFile(filepath.Join(dir, ".oxlintrc.json"))
	require.NoError(t, err)

	cfg, err := config.FromJSON(data)
	require.NoError(t, err, "written config must round-trip through config.FromJSON")
	require.NotEmpty(t, cfg.Rules, "generated config must enable rules")
}

func TestRepair_ExistingConfigIsUntouched(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, dir, "package.json", reactPackageJSON)
	writeFile(t, dir, ".oxlintrc.json", `{"rules":{}}`)

	result, err := provider.Provider.Repair.Repair(workingDirCtx(t, dir))

	require.NoError(t, err)
	require.Contains(t, result.Description, "already exists")

	data, err := os.ReadFile(filepath.Join(dir, ".oxlintrc.json"))
	require.NoError(t, err)
	require.JSONEq(t, `{"rules":{}}`, string(data), "repair must never overwrite an existing config")
}

func workingDirCtx(t *testing.T, dir string) context.Context {
	t.Helper()

	return finding.WithWorkingDir(context.Background(), dir)
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()

	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
}
