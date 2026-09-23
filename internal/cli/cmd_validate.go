package cli

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"

	autoconfigure "github.com/larsartmann/linter-autoconfigure-sdk"
	"github.com/larsartmann/oxlint-auto-configure/pkg/config"
	"github.com/larsartmann/oxlint-auto-configure/pkg/detect"
	"github.com/larsartmann/oxlint-auto-configure/pkg/diff"
	"github.com/larsartmann/oxlint-auto-configure/pkg/profile"
	"github.com/larsartmann/oxlint-auto-configure/pkg/rule"
	"github.com/spf13/cobra"
)

func newValidateCommand() *cobra.Command {
	var configPath string
	var failOnDrift bool

	cmd := &cobra.Command{
		Use:   CmdValidate,
		Short: "Validate an existing .oxlintrc.json configuration",
		RunE: func(_ *cobra.Command, _ []string) error {
			return Validate(configPath, failOnDrift)
		},
	}

	cmd.Flags().
		StringVarP(&configPath, "config", "c", "", "Config file path (default: .oxlintrc.json)")
	cmd.Flags().
		BoolVar(&failOnDrift, "fail-on-drift", false,
			"Exit non-zero when the config has drifted from what configure would generate (advisory by default)")

	return cmd
}

// Validate checks an existing .oxlintrc.json for unknown rules and invalid
// severities. Config I/O goes through linter-autoconfigure-sdk's typed
// LoadJSON so a missing or malformed file surfaces as a structured
// *ConfigError (errors.Is(err, fs.ErrNotExist) works for programmatic
// callers). After structural validation passes, the config is also compared
// against what configure would generate: drift is reported as an advisory
// warning (nothing is modified); pass failOnDrift to promote it to an error
// for CI gates.
func Validate(configPath string, failOnDrift bool) error {
	targetPath := configPath
	if targetPath == "" {
		targetPath = defaultConfigPath
	}

	cfg, cerr := autoconfigure.LoadJSON[config.OxlintConfig](targetPath)
	if cerr != nil {
		if errors.Is(cerr, fs.ErrNotExist) {
			return fmt.Errorf(
				"no config at %s; run `oxlint-auto-configure configure` to generate one: %w",
				targetPath, cerr)
		}

		return fmt.Errorf("load config %s: %w", targetPath, cerr)
	}

	reg, err := rule.LoadRegistry()
	if err != nil {
		return fmt.Errorf("load registry: %w", err)
	}

	result, err := config.ValidateConfig(cfg, reg)
	if err != nil {
		return fmt.Errorf("validate config: %w", err)
	}

	slog.Info(
		"config valid",
		"rules", len(cfg.Rules),
		"enabled", result.EnabledCount,
		"disabled", result.DisabledCount,
		"external", len(result.ExternalRules),
		"plugins", cfg.Plugins,
		"categories", cfg.Categories,
	)

	if reportDrift(targetPath, failOnDrift) {
		return fmt.Errorf(
			"config drift detected: %s has drifted from the generated config; "+
				"run `oxlint-auto-configure configure` to regenerate",
			targetPath)
	}

	return nil
}

// reportDrift compares the validated config against a fresh strict-profile
// generate run (the same pipeline `configure` uses, PreserveExternal applied
// so preserved user policy never reads as drift). Returns true only when
// drift was detected AND failOnDrift is set; the drift summary is logged as
// an advisory warning either way — nothing is modified. A failing drift
// check (e.g. unreadable package.json) degrades to a skip: structural
// validity already passed and the advisory must not fail the command.
func reportDrift(targetPath string, failOnDrift bool) bool {
	data, cerr := autoconfigure.ReadConfig(targetPath)
	if cerr != nil {
		slog.Warn("drift check skipped", "error", cerr)

		return false
	}

	existing, err := config.FromJSON(data)
	if err != nil {
		slog.Warn("drift check skipped", "error", err)

		return false
	}

	root := filepath.Dir(targetPath)

	reg, err := rule.LoadRegistry()
	if err != nil {
		slog.Warn("drift check skipped", "error", err)

		return false
	}

	det := detect.NewDetector(root)

	pluginConfig, projectTypes, err := det.Detect()
	if err != nil {
		slog.Warn("drift check skipped", "error", err)

		return false
	}

	expected, err := config.GenerateProjectConfig(
		profile.ProfileStrict, reg, pluginConfig, projectTypes, det.DetectExternalPlugins())
	if err != nil {
		slog.Warn("drift check skipped", "error", err)

		return false
	}

	differ := diff.NewDiffer(existing, config.PreserveExternal(existing, expected))
	if !differ.HasChanges() {
		return false
	}

	if failOnDrift {
		return true
	}

	slog.Warn("config drift detected (advisory — nothing was modified)",
		"summary", differ.Summary(),
		"fix", "run `oxlint-auto-configure configure` to regenerate",
	)

	return false
}
