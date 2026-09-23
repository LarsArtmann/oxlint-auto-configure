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
	require.NotNil(t, specs[0].HealthCheck, "provider must report config drift via its health check")
	require.Empty(t, specs[0].DependsOn, "provider must not depend on oxlint; oxlint depends on it")
	// package.json (the detector reads it) plus every config discovery name
	// the SDK derives from ProviderSpec.ConfigFiles — an existing curated
	// .oxlintrc.jsonc or oxlint.config.json affects this tool's findings too.
	require.Equal(t,
		[]string{"package.json", ".oxlintrc.json", ".oxlintrc.jsonc", "oxlint.config.json"},
		specs[0].Inputs)
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
	require.Equal(t, finding.FixStrategySuggest, findings[0].FixStrategy)
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

func TestDetect_ExistingJsoncConfigNeverFlagged(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, dir, "package.json", reactPackageJSON)
	writeFile(t, dir, ".oxlintrc.jsonc", "// curated config with inline rationale\n{\"rules\":{}}")

	findings, err := provider.Provider.Detect.Detect(workingDirCtx(t, dir))

	require.NoError(t, err)
	require.Empty(
		t,
		findings,
		"an existing .oxlintrc.jsonc is a config too — flagging it would make repair generate a shadowing .oxlintrc.json",
	)
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

// TestRepair_RegistersShadcnLintWithoutEnablingRules verifies the BuildFlow
// repair path follows the same external-plugin contract as the CLI: a
// project with the shadcn lint package installed gets it registered under
// "jsPlugins", while none of its design-system rules are turned on.
func TestRepair_RegistersShadcnLintWithoutEnablingRules(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, dir, "package.json", `{
		"name":"test-app",
		"dependencies":{"react":"^18.0.0"},
		"devDependencies":{"@shadcn/lint":"^1.0.0"}
	}`)

	_, err := provider.Provider.Repair.Repair(workingDirCtx(t, dir))
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dir, ".oxlintrc.json"))
	require.NoError(t, err)

	cfg, err := config.FromJSON(data)
	require.NoError(t, err)
	require.Equal(t, []string{"@shadcn/lint"}, cfg.JsPlugins)
	require.False(t, config.HasExternalRules(cfg),
		"repair registers the plugin but leaves design-system rules to the project")
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

func TestRepair_ExistingJsoncConfigIsNotShadowed(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, dir, "package.json", reactPackageJSON)
	writeFile(t, dir, ".oxlintrc.jsonc", "// curated config with inline rationale\n{\"rules\":{}}")

	result, err := provider.Provider.Repair.Repair(workingDirCtx(t, dir))

	require.NoError(t, err)
	require.Contains(t, result.Description, "already exists")

	jsonc, err := os.ReadFile(filepath.Join(dir, ".oxlintrc.jsonc"))
	require.NoError(t, err)
	require.Contains(t, string(jsonc), "curated config", "the curated .oxlintrc.jsonc must be untouched")
	require.NoFileExists(t, filepath.Join(dir, ".oxlintrc.json"),
		"generating a .oxlintrc.json next to a curated .oxlintrc.jsonc would shadow it — oxlint prefers .json")
}

func TestHealthCheck_NoConfigIsHealthy(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, dir, "package.json", reactPackageJSON)

	require.NoError(t, provider.Provider.HealthCheck(workingDirCtx(t, dir)),
		"a missing config is Detect's finding, not a health failure")
}

func TestHealthCheck_JsoncOnlyConfigIsHealthy(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, dir, "package.json", reactPackageJSON)
	writeFile(t, dir, ".oxlintrc.jsonc", `{"rules":{}}`)

	require.NoError(t, provider.Provider.HealthCheck(workingDirCtx(t, dir)),
		"a user-curated .oxlintrc.jsonc is out of scope: the tool never writes that file")
}

func TestHealthCheck_FreshRepairIsHealthy(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, dir, "package.json", reactPackageJSON)

	_, err := provider.Provider.Repair.Repair(workingDirCtx(t, dir))
	require.NoError(t, err)

	require.NoError(t, provider.Provider.HealthCheck(workingDirCtx(t, dir)),
		"a config this tool just wrote must match what it would generate")
}

func TestHealthCheck_DriftedConfigIsReported(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, dir, "package.json", reactPackageJSON)
	writeFile(t, dir, ".oxlintrc.json", `{"rules":{"no-console":"off"}}`)

	err := provider.Provider.HealthCheck(workingDirCtx(t, dir))

	require.Error(t, err)
	require.Contains(t, err.Error(), "drifted")
	require.Contains(t, err.Error(), "oxlint-auto-configure configure",
		"the error must name the fix")
	require.Contains(t, err.Error(), "advisory only",
		"the error must reassure that nothing was modified")

	data, err := os.ReadFile(filepath.Join(dir, ".oxlintrc.json"))
	require.NoError(t, err)
	require.JSONEq(t, `{"rules":{"no-console":"off"}}`, string(data),
		"the health check is report-only and must never touch the config")
}

func TestHealthCheck_MalformedConfigIsReported(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, dir, "package.json", reactPackageJSON)
	writeFile(t, dir, ".oxlintrc.json", `{"rules":`)

	err := provider.Provider.HealthCheck(workingDirCtx(t, dir))

	require.Error(t, err)
	require.Contains(t, err.Error(), "malformed")
	require.Contains(t, err.Error(), "configure", "the error must name the fix")
}

func TestHealthCheck_PreservedOverridesAreNotDrift(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, dir, "package.json", reactPackageJSON)

	_, err := provider.Provider.Repair.Repair(workingDirCtx(t, dir))
	require.NoError(t, err)

	// Simulate a project adding an @shadcn/lint-style overrides block to the
	// generated config: regeneration (and therefore the drift check) must
	// treat the block as preserved policy, not drift.
	path := filepath.Join(dir, ".oxlintrc.json")
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	cfg, err := config.FromJSON(data)
	require.NoError(t, err)

	cfg.Overrides = append(cfg.Overrides, map[string]any{
		"files": []any{"src/components/ui/**"},
		"rules": map[string]any{"shadcn/no-restyle": "off"},
	})
	updated, err := cfg.ToJSON()
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, updated, 0o600))

	require.NoError(t, provider.Provider.HealthCheck(workingDirCtx(t, dir)),
		"overrides are preserved verbatim on regeneration, so they cannot be drift")
}

func workingDirCtx(t *testing.T, dir string) context.Context {
	t.Helper()

	return finding.WithWorkingDir(context.Background(), dir)
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()

	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
}
