package cli

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"

	autoconfigure "github.com/larsartmann/linter-autoconfigure-sdk"
	"github.com/larsartmann/oxlint-auto-configure/pkg/config"
	"github.com/larsartmann/oxlint-auto-configure/pkg/rule"
	"github.com/spf13/cobra"
)

func newValidateCommand() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   CmdValidate,
		Short: "Validate an existing .oxlintrc.json configuration",
		RunE: func(_ *cobra.Command, _ []string) error {
			return Validate(configPath)
		},
	}

	cmd.Flags().
		StringVarP(&configPath, "config", "c", "", "Config file path (default: .oxlintrc.json)")

	return cmd
}

// Validate checks an existing .oxlintrc.json for unknown rules and invalid severities.
// Config I/O goes through linter-autoconfigure-sdk's typed LoadJSON so a
// missing or malformed file surfaces as a structured *ConfigError
// (errors.Is(err, fs.ErrNotExist) works for programmatic callers).
func Validate(configPath string) error {
	targetPath := configPath
	if targetPath == "" {
		targetPath = defaultConfigPath
	}

	cfg, cerr := autoconfigure.LoadJSON[config.OxlintConfig](targetPath)
	if cerr != nil {
		if errors.Is(cerr, fs.ErrNotExist) {
			return fmt.Errorf(
				"no config at %s; run `oxlint-auto-configure configure` to generate one",
				targetPath)
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

	return nil
}
