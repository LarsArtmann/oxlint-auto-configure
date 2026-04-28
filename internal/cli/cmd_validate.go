package cli

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/larsartmann/oxlint-auto-configure/pkg/config"
	"github.com/larsartmann/oxlint-auto-configure/pkg/rule"
	"github.com/spf13/cobra"
)

func newValidateCommand() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate an existing .oxlintrc.json configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			var unknown []string
			for name := range cfg.Rules {
				if _, ok := reg.ByName(name); !ok {
					unknown = append(unknown, name)
				}
			}

			if len(unknown) > 0 {
				slog.Error("unknown rules", "rules", strings.Join(unknown, ", "))
				return fmt.Errorf("%d unknown rules found", len(unknown))
			}

			validSeverities := map[string]bool{"error": true, "warn": true, "off": true}
			var invalid []string
			for name, sev := range cfg.Rules {
				if !validSeverities[sev] {
					invalid = append(invalid, fmt.Sprintf("%s=%s", name, sev))
				}
			}

			if len(invalid) > 0 {
				slog.Error("invalid severities", "rules", strings.Join(invalid, ", "))
				return fmt.Errorf("%d invalid severities found", len(invalid))
			}

			enabled := 0
			for _, sev := range cfg.Rules {
				if sev != "off" {
					enabled++
				}
			}

			slog.Info("config valid",
				"path", targetPath,
				"rules", len(cfg.Rules),
				"enabled", enabled,
				"disabled", len(cfg.Rules)-enabled,
				"plugins", cfg.Plugins,
				"categories", cfg.Categories,
			)

			return nil
		},
	}

	cmd.Flags().
		StringVarP(&configPath, "config", "c", "", "Config file path (default: .oxlintrc.json)")

	return cmd
}
