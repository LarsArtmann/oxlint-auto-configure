// Package provider wires oxlint-auto-configure into BuildFlow's DAG via the
// go-finding/toolsdk Spec contract. BuildFlow discovers this Provider
// automatically through toolsdk.All() when a consumer blank-imports this
// package:
//
//	import _ "github.com/larsartmann/oxlint-auto-configure/pkg/provider"
//
// The Spec is built with linter-autoconfigure-sdk's BootstrapProviderFromSpec
// so the generate-if-missing lifecycle (missing-only Detect, never-overwrite
// dry-run-aware Repair, advisory drift HealthCheck) is owned by the shared
// SDK; this package supplies only the oxlint domain: what "recognizable
// project" means, how to generate the config, and how to compare two.
package provider

import (
	"context"
	"fmt"

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

// oxlintConfigFiles are the config file names oxlint itself searches for, in
// priority order (mirrors BuildFlow's findOxlintConfig in
// tools/providers/js_tools.go). All of them count as "a config exists":
// generating a fresh .oxlintrc.json in a repo whose curated config is
// .oxlintrc.jsonc would SHADOW it (oxlint prefers .json), silently
// clobbering the effective configuration — the exact overwrite the
// never-overwrite contract forbids.
//
//nolint:gochecknoglobals // mirrors BuildFlow's package-level oxlintConfigFiles
var oxlintConfigFiles = []string{
	".oxlintrc.json",
	".oxlintrc.jsonc",
	"oxlint.config.json",
}

// oxlintConfigFilePaths is oxlintConfigFiles as the SDK's branded type for
// BootstrapSpec.ConfigFiles.
func oxlintConfigFilePaths() []finding.FilePath {
	paths := make([]finding.FilePath, 0, len(oxlintConfigFiles))
	for _, name := range oxlintConfigFiles {
		paths = append(paths, finding.FilePath(name))
	}

	return paths
}

//nolint:gochecknoglobals // BuildFlow plugin SDK requires package-level Provider registration
var Provider = mustProvider()

// mustProvider builds the Spec through the shared SDK bootstrap bridge and
// layers the oxlint-specific DAG wiring (Trigger, DependsOn) on top.
// Registration panics on an invalid Spec, the same contract as
// toolsdk.Register.
func mustProvider() toolsdk.Spec {
	spec, err := autoconfigure.BootstrapProviderFromSpec(
		autoconfigure.BootstrapSpec[*config.OxlintConfig]{
			Name:        toolName,
			Description: "Detects a missing " + configFileName + " and repairs it: generates the optimal " + configFileName + " for the detected project type (React, Next.js, Vue, ...)",
			ConfigFile:  finding.FilePath(configFileName),
			ConfigFiles: oxlintConfigFilePaths(),
			MissingRule: missingConfigRule,
			FixCommand:  "oxlint-auto-configure configure",
			CountLabel:  "rules",
			Recognizable: func(ctx context.Context) (bool, error) {
				_, projectTypes, detectErr := detect.NewDetector(autoconfigure.WorkingDir(ctx)).Detect()
				if detectErr != nil {
					return false, fmt.Errorf("%s detect: %w", toolName, detectErr)
				}

				return hasKnownProjectType(projectTypes), nil
			},
			Generate:          generateConfig,
			Marshal:           marshalConfig,
			Parse:             parseConfig,
			NormalizeExpected: config.PreserveExternal,
			Compare: func(existing, expected *config.OxlintConfig) []autoconfigure.Change {
				return diff.NewDiffer(existing, expected).Diff()
			},
		},
	)
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
	// The SDK derives Inputs from ConfigFiles (all three discovery names — an
	// existing curated .oxlintrc.jsonc or oxlint.config.json also affects
	// this tool's findings); the detector reads package.json on top, so
	// prepend it to the derived read contract.
	spec.Inputs = append([]string{"package.json"}, spec.Inputs...)

	return toolsdk.Register(spec)
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

// generateConfig runs the detect → profile → generate pipeline shared with
// the CLI's configure command and returns the config plus its rule count.
// The strict profile is the DAG default; custom profiles remain a CLI
// concern.
func generateConfig(ctx context.Context) (*config.OxlintConfig, int, error) {
	root := autoconfigure.WorkingDir(ctx)

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

	return cfg, len(cfg.Rules), nil
}

// marshalConfig renders the config to the exact file bytes ToJSON produces
// plus the trailing newline the tool's file format carries.
func marshalConfig(cfg *config.OxlintConfig) ([]byte, *autoconfigure.ConfigError) {
	data, err := cfg.ToJSON()
	if err != nil {
		return nil, &autoconfigure.ConfigError{
			Op:   autoconfigure.OpMarshal,
			Path: configFileName,
			Err:  fmt.Errorf("%s marshal: %w", toolName, err),
		}
	}

	return append(data, '\n'), nil
}

// parseConfig reads config bytes through the domain's FromJSON, surfaced as
// the SDK's typed config error.
func parseConfig(data []byte) (*config.OxlintConfig, *autoconfigure.ConfigError) {
	cfg, err := config.FromJSON(data)
	if err != nil {
		return nil, &autoconfigure.ConfigError{
			Op:   autoconfigure.OpUnmarshal,
			Path: configFileName,
			Err:  fmt.Errorf("%s unmarshal: %w", toolName, err),
		}
	}

	return cfg, nil
}
