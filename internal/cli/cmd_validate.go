package cli

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/larsartmann/oxlint-auto-configure/pkg/config"
	"github.com/larsartmann/oxlint-auto-configure/pkg/rule"
	"github.com/spf13/cobra"
)

func newValidateCommand() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "validate",
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
func Validate(configPath string) error {
	targetPath := configPath
	if targetPath == "" {
		targetPath = defaultConfigPath
	}

	data, err := os.ReadFile(targetPath)
	if err != nil {
		return fmt.Errorf("read config %s: %w", targetPath, err)
	}

	cfg, err := config.FromJSON(data)
	if err != nil {
		return fmt.Errorf("parse config: %w", err)
	}

	reg, err := rule.LoadRegistry()
	if err != nil {
		return fmt.Errorf("load registry: %w", err)
	}

	result, err := config.ValidateConfig(cfg, reg)
	if err != nil {
		return fmt.Errorf("validate config: %w", err)
	}

	slog.Info("config valid",
		"rules", len(cfg.Rules),
		"enabled", result.EnabledCount,
		"disabled", result.DisabledCount,
		"plugins", cfg.Plugins,
		"categories", cfg.Categories,
	)

	return nil
}
