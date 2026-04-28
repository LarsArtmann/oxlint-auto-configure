package cli

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/larsartmann/oxlint-auto-configure/pkg/config"
	"github.com/larsartmann/oxlint-auto-configure/pkg/detect"
	"github.com/larsartmann/oxlint-auto-configure/pkg/diff"
	"github.com/larsartmann/oxlint-auto-configure/pkg/oxlint"
	"github.com/larsartmann/oxlint-auto-configure/pkg/profile"
	"github.com/larsartmann/oxlint-auto-configure/pkg/rule"
	"github.com/spf13/cobra"
)

func newConfigureCommand() *cobra.Command {
	var (
		profileFlag string
		configPath  string
		dryRun      bool
		runFix      bool
		rootDir     string
	)

	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Generate optimal .oxlintrc.json configuration",
		Long: `Analyze the project and generate the best oxlint configuration.

Profiles:
  maximal-typesafe  Enable ALL rules at 'error' — maximum type safety
  recommended       Correctness+suspicious+TS at error, rest at warn (default)
  strict            Correctness+suspicious at error, everything else at warn
  minimal           Only correctness at error, rest uses oxlint defaults`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if rootDir == "" {
				rootDir = "."
			}

			absRoot, err := filepath.Abs(rootDir)
			if err != nil {
				return fmt.Errorf("resolve root dir: %w", err)
			}

			p := profile.Profile(profileFlag)
			if !p.IsValid() {
				return fmt.Errorf(
					"invalid profile %q: choose from %s",
					profileFlag,
					strings.Join(profileNames(), ", "),
				)
			}

			oxlintVer, err := oxlint.CheckVersion(cmd.Context())
			if err != nil {
				return fmt.Errorf("oxlint: %w", err)
			}
			slog.Info("oxlint version", "version", oxlintVer)

			embeddedVer := rule.EmbeddedVersion()
			if embeddedVer != "" && embeddedVer != oxlintVer {
				slog.Warn("embedded rules version mismatch",
					"embedded", embeddedVer,
					"runtime", oxlintVer,
				)
			}

			reg, err := rule.LoadRegistry()
			if err != nil {
				return fmt.Errorf("load rule registry: %w", err)
			}

			det := detect.NewDetector(absRoot)
			pluginConfig, projectTypes, err := det.Detect()
			if err != nil {
				return fmt.Errorf("detect project type: %w", err)
			}

			slog.Info("detected project", "types", detect.FormatTypes(projectTypes))
			slog.Info("profile", "name", p)
			slog.Info("rules loaded", "total", reg.Len())

			cat := profile.NewCategorizer(p, pluginConfig)
			gen := config.NewGenerator(cat, reg)

			var cfg *config.OxlintConfig
			if p == profile.ProfileMaximalTypesafe {
				cfg = gen.GenerateAllError()
			} else {
				cfg = gen.Generate()
			}

			targetPath := configPath
			if targetPath == "" {
				targetPath = filepath.Join(absRoot, defaultConfigPath)
			}

			if dryRun {
				return writeDryRun(cfg, targetPath)
			}

			if existingData, err := os.ReadFile(targetPath); err == nil {
				existing, err := config.FromJSON(existingData)
				if err == nil {
					d := diff.NewDiffer(existing, cfg)
					slog.Info("changes\n" + d.FormatDiff())
					slog.Info(d.Summary())
				}
			}

			data, err := cfg.ToJSON()
			if err != nil {
				return fmt.Errorf("generate config JSON: %w", err)
			}

			if err := os.WriteFile(targetPath, append(data, '\n'), 0o644); err != nil {
				return fmt.Errorf("write config: %w", err)
			}

			slog.Info("configuration written", "path", targetPath)

			if runFix {
				slog.Info("running oxlint fix")
				fixResult, err := oxlint.RunFix(cmd.Context(), absRoot, targetPath)
				if err != nil {
					return fmt.Errorf("fix: %w", err)
				}
				if fixResult.Output != "" {
					slog.Info("fix output", "detail", fixResult.Output)
				}
				slog.Info("fix complete")
			}

			return nil
		},
	}

	cmd.Flags().
		StringVarP(&profileFlag, "profile", "p", string(defaultProfile), "Configuration profile")
	cmd.Flags().
		StringVarP(&configPath, "config", "c", "", "Output config file path (default: .oxlintrc.json)")
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Show what would change without writing")
	cmd.Flags().BoolVar(&runFix, "fix", false, "Run oxlint --fix after writing config")
	cmd.Flags().StringVar(&rootDir, "root", ".", "Project root directory")

	return cmd
}

func writeDryRun(cfg *config.OxlintConfig, targetPath string) error {
	data, err := cfg.ToJSON()
	if err != nil {
		return fmt.Errorf("generate config JSON: %w", err)
	}

	slog.Info("dry run", "path", targetPath)
	fmt.Println(string(data))
	return nil
}

func profileNames() []string {
	ps := profile.AllProfiles()
	names := make([]string, len(ps))
	for i, p := range ps {
		names[i] = string(p)
	}
	return names
}
