// Package provider wires oxlint-auto-configure into BuildFlow's DAG via the
// go-finding/toolsdk Spec contract. BuildFlow discovers this Provider
// automatically through toolsdk.All() when a consumer blank-imports this
// package:
//
//	import _ "github.com/larsartmann/oxlint-auto-configure/pkg/provider"
//
// The Spec is built with linter-autoconfigure-sdk's ProviderFromSpec so the
// domain shape (Analyze reporting ConfigIssues, Repair describing what was
// rewritten) stays in the auto-configurer's language and the finding emission
// is owned by the shared SDK instead of re-implemented here.
package provider

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	atomicwrite "github.com/larsartmann/go-atomic-write"
	"github.com/larsartmann/go-finding"
	toolsdk "github.com/larsartmann/go-finding/toolsdk"
	autoconfigure "github.com/larsartmann/linter-autoconfigure-sdk"
	"github.com/larsartmann/oxlint-auto-configure/pkg/config"
	"github.com/larsartmann/oxlint-auto-configure/pkg/detect"
	"github.com/larsartmann/oxlint-auto-configure/pkg/diff"
	"github.com/larsartmann/oxlint-auto-configure/pkg/profile"
	"github.com/larsartmann/oxlint-auto-configure/pkg/rule"
)

const (
	// toolName is the BuildFlow DAG tool name. It must stay in sync with
	// BuildFlow's config.ToolOxlintAutoConfigure.
	toolName = "oxlint-auto-configure"

	// configFileName is the config file this tool detects and generates.
	configFileName = ".oxlintrc.json"

	// missingConfigRule is the go-finding rule name reported when a
	// recognizable JS/TS project has no oxlint config yet.
	missingConfigRule = "OXLOPT_CONFIG_MISSING"
)

// errConfigDrift is wrapped by the HealthCheck when an existing config no
// longer matches what the tool would generate. Advisory only: consumers
// must never treat it as a trigger to rewrite the config.
var errConfigDrift = errors.New("config drift detected")

// oxlintConfigFiles are the config file names oxlint itself searches for, in
// priority order (mirrors BuildFlow's findOxlintConfig in
// tools/providers/js_tools.go). hasConfig must recognize ALL of them:
// generating a fresh .oxlintrc.json in a repo whose curated config is
// .oxlintrc.jsonc would SHADOW it (oxlint prefers .json), silently
// clobbering the effective configuration — the exact overwrite the
// "never overwrite an existing config" contract forbids.
//
//nolint:gochecknoglobals // mirrors BuildFlow's package-level oxlintConfigFiles
var oxlintConfigFiles = []string{
	".oxlintrc.json",
	".oxlintrc.jsonc",
	"oxlint.config.json",
}

// workingDir resolves the project directory from the context, falling back
// to the process working directory, mirroring the other BuildFlow providers
// so the WithWorkingDir fan-out works identically.
func workingDir(ctx context.Context) string {
	if dir := finding.WorkingDirFromContext(ctx); dir != "" {
		return dir
	}

	return "."
}

// configPath returns the absolute path of the oxlint config inside root.
func configPath(root string) string {
	return filepath.Join(root, configFileName)
}

// hasConfig reports whether any oxlint config file already exists in root.
// All three discovery names count: an existing curated config must suppress
// regeneration regardless of which filename it uses.
func hasConfig(root string) bool {
	for _, name := range oxlintConfigFiles {
		if _, err := os.Stat(filepath.Join(root, name)); err == nil {
			return true
		}
	}

	return false
}

//nolint:gochecknoglobals // BuildFlow plugin SDK requires package-level Provider registration
var Provider = mustProvider()

// mustProvider builds the Spec through the shared SDK bridge and layers the
// oxlint-specific DAG wiring (Trigger, DependsOn) on top. Registration
// panics on an invalid Spec, the same contract as toolsdk.Register.
func mustProvider() toolsdk.Spec {
	spec, err := autoconfigure.ProviderFromSpec(autoconfigure.ProviderSpec{
		Name:        toolName,
		Description: "Detects a missing " + configFileName + " and repairs it: generates the optimal " + configFileName + " for the detected project type (React, Next.js, Vue, ...)",
		ConfigFile:  finding.FilePath(configFileName),
		Analyze:     detectMissingConfig,
		Repair:      repairConfig,
	})
	if err != nil {
		panic("provider: invalid spec: " + err.Error())
	}

	spec.Trigger = toolsdk.OnFiles(
		"javascript",
		"**/*.js", "**/*.jsx", "**/*.ts", "**/*.tsx", "**/*.mjs", "**/*.cjs",
		"**/*.vue", "**/*.svelte", "**/*.astro",
	)
	// nil on purpose (DAG flip 2026-09-11, owner decision): the oxlint LINT
	// step depends on US (BuildFlow's NewOxlintProvider WithDeps), so a
	// missing config is generated before linting in the same run.
	spec.DependsOn = nil
	// ProviderFromSpec derives Inputs from ConfigFile alone; the detector
	// also reads package.json, so restore the full read contract.
	spec.Inputs = []string{"package.json", configFileName}
	// HealthCheck reports config drift as an advisory. BuildFlow treats a
	// failing health check as report-only (warn log + summary entry): it
	// never skips the tool and never triggers a repair, so flagging drift
	// here cannot cause Repair to stomp user customizations — Detect stays
	// missing-only and Repair keeps its never-overwrite contract.
	spec.HealthCheck = healthCheckDrift

	return toolsdk.Register(spec)
}

// detectMissingConfig reports a warning finding when the project is a
// recognizable JS/TS project without an oxlint config. An existing config is
// never flagged: this tool generates a fresh optimal config, so flagging a
// user-customized config would lead Repair to stomp it.
func detectMissingConfig(ctx context.Context) ([]autoconfigure.ConfigIssue, error) {
	root := workingDir(ctx)
	if hasConfig(root) {
		return nil, nil
	}

	_, projectTypes, err := detect.NewDetector(root).Detect()
	if err != nil {
		return nil, fmt.Errorf("%s detect: %w", toolName, err)
	}

	if !hasKnownProjectType(projectTypes) {
		return nil, nil
	}

	return []autoconfigure.ConfigIssue{
		{
			Rule:        missingConfigRule,
			Message:     "No " + configFileName + " found; generate the optimal config for the detected project type",
			Severity:    finding.SeverityWarning,
			File:        finding.FilePath(configFileName),
			Line:        0,
			Suggestion:  "run `oxlint-auto-configure configure` or apply this repair to write " + configFileName,
			Confidence:  finding.ConfidenceHigh,
			FixStrategy: nil,
		},
	}, nil
}

// healthCheckDrift is the toolsdk Spec's HealthCheck: it verifies that an
// existing .oxlintrc.json still matches what the tool would generate for the
// detected project type. A missing config is healthy (Detect owns the
// missing-config finding, and Repair generates on demand); the other config
// file names oxlint accepts (.oxlintrc.jsonc, oxlint.config.json) are
// user-curated configs this tool never writes, so they are out of scope. A
// returned error is advisory by contract: it wraps errConfigDrift, names the
// fix, and states that nothing was modified.
func healthCheckDrift(ctx context.Context) error {
	root := workingDir(ctx)

	path := configPath(root)
	if _, err := os.Stat(path); err != nil {
		//nolint:nilerr // a missing config is healthy: Detect owns that finding
		return nil
	}

	existingData, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("%s health: read %s: %w", toolName, configFileName, err)
	}

	existing, err := config.FromJSON(existingData)
	if err != nil {
		return fmt.Errorf(
			"%s health: %s is malformed (%w); run `oxlint-auto-configure configure` to regenerate",
			toolName, configFileName, err)
	}

	expectedData, _, err := generateConfig(root)
	if err != nil {
		return fmt.Errorf("%s health: generate expected config: %w", toolName, err)
	}

	expected, err := config.FromJSON(expectedData)
	if err != nil {
		return fmt.Errorf("%s health: re-parse generated config: %w", toolName, err)
	}

	// Compare against what configure would actually write: the generated
	// config plus every external-plugin bit carried over from the existing
	// file. Without this, a deliberately customized external setup would
	// read as drift on every run.
	expected = config.PreserveExternal(existing, expected)

	d := diff.NewDiffer(existing, expected)
	if !d.HasChanges() {
		return nil
	}

	return fmt.Errorf(
		"%w: %s has drifted from the generated config (%s); "+
			"run `oxlint-auto-configure configure` to regenerate "+
			"(advisory only — nothing was modified)",
		errConfigDrift, configFileName, d.Summary())
}

// hasKnownProjectType reports whether at least one detected project type is
// recognized (i.e. the project is not an unidentifiable directory).
func hasKnownProjectType(projectTypes []detect.ProjectType) bool {
	for _, pt := range projectTypes {
		if pt != detect.ProjectTypeUnknown {
			return true
		}
	}

	return false
}

// repairConfig generates the optimal oxlint config for the detected project
// type and writes it, honoring the dry-run flag carried by the context. An
// existing config is never overwritten: Repair only fires for projects whose
// config is missing.
func repairConfig(ctx context.Context) (string, error) {
	root := workingDir(ctx)
	dryRun := toolsdk.DryRunFromContext(ctx)

	if hasConfig(root) {
		return configFileName + " already exists; keeping the existing configuration", nil
	}

	data, ruleCount, err := generateConfig(root)
	if err != nil {
		return "", err
	}

	if dryRun {
		return fmt.Sprintf("held back by dry-run: would write %s (%d rules)", configFileName, ruleCount), nil
	}

	if err := writeConfigFile(configPath(root), data); err != nil {
		return "", err
	}

	return fmt.Sprintf("wrote %s (%d rules)", configFileName, ruleCount), nil
}

// generateConfig runs the detect → profile → generate pipeline shared with
// the CLI's configure command and returns the marshaled config plus its rule
// count. The strict profile is the DAG default; custom profiles remain a
// CLI concern.
func generateConfig(root string) ([]byte, int, error) {
	reg, err := rule.LoadRegistry()
	if err != nil {
		return nil, 0, fmt.Errorf("%s repair: load rule registry: %w", toolName, err)
	}

	det := detect.NewDetector(root)

	pluginConfig, projectTypes, err := det.Detect()
	if err != nil {
		return nil, 0, fmt.Errorf("%s repair: detect project type: %w", toolName, err)
	}

	externalPlugins := det.DetectExternalPlugins()

	cfg, err := config.GenerateProjectConfig(
		profile.ProfileStrict, reg, pluginConfig, projectTypes, externalPlugins)
	if err != nil {
		return nil, 0, fmt.Errorf("%s repair: generate config: %w", toolName, err)
	}

	data, err := cfg.ToJSON()
	if err != nil {
		return nil, 0, fmt.Errorf("%s repair: marshal config: %w", toolName, err)
	}

	return data, len(cfg.Rules), nil
}

// writeConfigFile writes the config atomically so a crash mid-write cannot
// truncate an existing file.
func writeConfigFile(path string, data []byte) error {
	if err := atomicwrite.Write(path, append(data, '\n')); err != nil {
		return fmt.Errorf("%s repair: write %s: %w", toolName, configFileName, err)
	}

	return nil
}
