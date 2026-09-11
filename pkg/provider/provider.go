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
	"fmt"
	"os"
	"path/filepath"

	atomicwrite "github.com/larsartmann/go-atomic-write"
	"github.com/larsartmann/go-finding"
	toolsdk "github.com/larsartmann/go-finding/toolsdk"
	autoconfigure "github.com/larsartmann/linter-autoconfigure-sdk"
	"github.com/larsartmann/oxlint-auto-configure/pkg/config"
	"github.com/larsartmann/oxlint-auto-configure/pkg/detect"
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

// hasConfig reports whether an oxlint config already exists in root.
func hasConfig(root string) bool {
	_, err := os.Stat(configPath(root))

	return err == nil
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
	spec.DependsOn = []string{"oxlint"}
	// nil = always healthy: config generation uses the embedded rule registry
	// and never shells out to the oxlint binary.
	spec.HealthCheck = nil

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
// count. The recommended profile is the DAG default; custom profiles remain a
// CLI concern.
func generateConfig(root string) ([]byte, int, error) {
	reg, err := rule.LoadRegistry()
	if err != nil {
		return nil, 0, fmt.Errorf("%s repair: load rule registry: %w", toolName, err)
	}

	pluginConfig, projectTypes, err := detect.NewDetector(root).Detect()
	if err != nil {
		return nil, 0, fmt.Errorf("%s repair: detect project type: %w", toolName, err)
	}

	cfg, err := config.GenerateProjectConfig(profile.ProfileRecommended, reg, pluginConfig, projectTypes)
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
