package cli

import (
	"context"
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			if rootDir == "" {
				rootDir = "."
			}
			absRoot, err := filepath.Abs(rootDir)
			if err != nil {
				return fmt.Errorf("resolve root dir: %w", err)
			}

			opts := ConfigureOptions{
				Profile:    profile.Profile(profileFlag),
				ConfigPath: configPath,
				DryRun:     dryRun,
				Fix:        runFix,
			}
			return Configure(cmd.Context(), absRoot, opts)
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

// ConfigureOptions holds the parameters for the configure operation.
type ConfigureOptions struct {
	Profile    profile.Profile
	ConfigPath string
	DryRun     bool
	Fix        bool
}

// Configure generates an oxlint configuration for the project at absRoot.
func Configure(ctx context.Context, absRoot string, opts ConfigureOptions) error {
	if !opts.Profile.IsValid() {
		return fmt.Errorf(
			"invalid profile %q: choose from %s",
			opts.Profile,
			strings.Join(profileNames(), ", "),
		)
	}

	if err := checkOxlintVersion(ctx); err != nil {
		return err
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
	slog.Info("profile", "name", opts.Profile)
	slog.Info("rules loaded", "total", reg.Len())

	cat := profile.NewCategorizer(opts.Profile, pluginConfig)
	gen := config.NewGenerator(cat, reg)

	cfg := genConfig(gen, opts.Profile)
	targetPath := resolveConfigPath(opts.ConfigPath, absRoot)

	if opts.DryRun {
		return writeDryRun(cfg, targetPath)
	}

	logDiffIfExisting(targetPath, cfg)

	if err := writeConfig(cfg, targetPath); err != nil {
		return err
	}

	return runFixIfNeeded(ctx, absRoot, targetPath, opts.Fix)
}

func checkOxlintVersion(ctx context.Context) error {
	oxlintVer, err := oxlint.CheckVersion(ctx)
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

	return nil
}

func genConfig(gen *config.Generator, p profile.Profile) *config.OxlintConfig {
	if p == profile.ProfileMaximalTypesafe {
		return gen.GenerateMaximal()
	}
	return gen.Generate()
}

func resolveConfigPath(configPath, absRoot string) string {
	if configPath == "" {
		return filepath.Join(absRoot, defaultConfigPath)
	}
	return configPath
}

func logDiffIfExisting(targetPath string, cfg *config.OxlintConfig) {
	if diff := showDiffIfExisting(targetPath, cfg); diff != "" {
		slog.Info(diff)
	}
}

func writeConfig(cfg *config.OxlintConfig, targetPath string) error {
	data, err := cfg.ToJSON()
	if err != nil {
		return fmt.Errorf("generate config JSON: %w", err)
	}

	if err := os.WriteFile(targetPath, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	slog.Info("configuration written", "path", targetPath)

	return nil
}

func runFixIfNeeded(ctx context.Context, absRoot, targetPath string, runFix bool) error {
	if !runFix {
		return nil
	}

	slog.Info("running oxlint fix")
	fixResult, err := oxlint.RunFix(ctx, absRoot, targetPath)
	if err != nil {
		return fmt.Errorf("fix: %w", err)
	}
	if fixResult.Output != "" {
		slog.Info("fix output", "detail", fixResult.Output)
	}
	slog.Info("fix complete")

	return nil
}

// showDiffIfExisting compares the existing config with the new one and logs the diff.
func showDiffIfExisting(targetPath string, cfg *config.OxlintConfig) string {
	existingData, err := os.ReadFile(targetPath)
	if err != nil {
		return ""
	}
	existing, err := config.FromJSON(existingData)
	if err != nil {
		slog.Warn("existing config is malformed, skipping diff", "error", err)

		return ""
	}
	d := diff.NewDiffer(existing, cfg)

	return d.Summary() + "\n" + d.FormatDiff()
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
