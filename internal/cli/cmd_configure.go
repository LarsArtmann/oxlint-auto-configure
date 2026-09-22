package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	atomicwrite "github.com/larsartmann/go-atomic-write"
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
		configPath  string
		dryRun      bool
		runFix      bool
		rootDir     string
		profileFlag string
	)

	cmd := &cobra.Command{
		Use:   CmdConfigure,
		Short: "Generate optimal .oxlintrc.json configuration",
		Long: `Analyze the project and generate the best oxlint configuration.

Profiles:
  maximal-typesafe  Enable ALL rules at 'error' — maximum type safety
  recommended       Correctness+suspicious at error, rest at warn (default)
  strict            Correctness+suspicious at error, everything else at warn
  minimal           Only correctness at error, rest uses oxlint defaults`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if rootDir == "" {
				rootDir = "."
			}

			absRoot, err := filepath.Abs(rootDir)
			if err != nil {
				return fmt.Errorf("profileFlag=%s: resolve root dir: %w", profileFlag, err)
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

	AddProfileFlag(cmd, &profileFlag)
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
			"%w %q: choose from %s",
			config.ErrInvalidProfile,
			opts.Profile,
			strings.Join(profile.AllProfileNames(), ", "),
		)
	}

	oxlintVer, oxlintFound, err := checkOxlintVersion(ctx)
	if err != nil {
		return err
	}


	reg, err := rule.LoadRegistry()
	if err != nil {
		return fmt.Errorf("absRoot=%s: load rule registry: %w", absRoot, err)
	}

	det := detect.NewDetector(absRoot)

	pluginConfig, projectTypes, err := det.Detect()
	if err != nil {
		return fmt.Errorf("absRoot=%s: detect project type: %w", absRoot, err)
	}

	externalPlugins := det.DetectExternalPlugins()

	slog.Info("detected project", "types", detect.FormatTypes(projectTypes))
	slog.Info("profile", "name", opts.Profile)
	slog.Info("rules loaded", "total", reg.Len())

	cfg, err := config.GenerateProjectConfig(opts.Profile, reg, pluginConfig, projectTypes, externalPlugins)
	if err != nil {
		return fmt.Errorf("absRoot=%s: generate config: %w", absRoot, err)
	}

	targetPath := resolveConfigPath(opts.ConfigPath, absRoot)

	cfg, preserved := preserveExistingExternal(targetPath, cfg)
	if preserved > 0 {
		slog.Info("preserved external plugins from existing config", "count", preserved)
	}

	warnOrphanedJsPlugins(cfg, externalPlugins)
	warnJsPluginsVersion(oxlintFound, oxlintVer, cfg)
	logExternalPluginHint(cfg, externalPlugins)

	if opts.DryRun {
		return writeDryRun(cfg, targetPath)
	}

	logDiffIfExisting(targetPath, cfg)

	if err := writeConfig(cfg, targetPath); err != nil {
		return err
	}

	return runFixIfNeeded(ctx, absRoot, targetPath, opts.Fix)
}

// checkOxlintVersion runs `oxlint --version` once per Configure run and
// reports both the version and whether a binary was found at all. A missing
// binary is a warning — generation works from the embedded registry — while
// any other failure aborts the run.
func checkOxlintVersion(ctx context.Context) (version string, found bool, err error) {
	oxlintVer, err := oxlint.CheckVersion(ctx)
	if err != nil {
		if errors.Is(err, oxlint.ErrNotFound) {
			slog.Warn("oxlint not found in PATH; skipping version check")

			return "", false, nil
		}

		return "", false, fmt.Errorf("oxlint version check: %w", err)
	}

	slog.Info("oxlint version", "version", oxlintVer)

	embeddedVer := rule.EmbeddedVersion()
	if embeddedVer != "" && embeddedVer != oxlintVer {
		slog.Warn(
			"embedded rules version mismatch",
			"embedded", embeddedVer,
			"runtime", oxlintVer,
		)
	}

	return oxlintVer, true, nil
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

// preserveExistingExternal loads the existing config at targetPath (if any)
// and carries its external-plugin registrations, rules, and settings into
// the freshly generated config so regeneration cannot destroy setup the
// tool cannot re-derive (e.g. an @shadcn/lint design-system policy). The
// second return value is the number of jsPlugins registrations carried over.
func preserveExistingExternal(targetPath string, cfg *config.OxlintConfig) (*config.OxlintConfig, int) {
	existingData, err := os.ReadFile(targetPath)
	if err != nil {
		return cfg, 0
	}

	existing, err := config.FromJSON(existingData)
	if err != nil {
		slog.Warn("existing config is malformed, cannot preserve external plugins", "error", err)

		return cfg, 0
	}

	return config.PreserveExternal(existing, cfg), len(existing.JsPlugins)
}

// warnJsPluginsVersion warns when the config registers external JS plugins
// but the oxlint binary found earlier in this run is older than the first
// version that supports the "jsPlugins" key. A missing binary is not a
// warning: version checks already skip it elsewhere.
func warnJsPluginsVersion(oxlintFound bool, oxlintVer string, cfg *config.OxlintConfig) {
	if len(cfg.JsPlugins) == 0 || !oxlintFound {
		return
	}

	if !oxlint.VersionAtLeast(oxlintVer, oxlint.MinVersionForJsPlugins) {
		slog.Warn("registered jsPlugins require a newer oxlint",
			"found", oxlintVer,
			"required", oxlint.MinVersionForJsPlugins,
		)
	}
}

// warnOrphanedJsPlugins warns about a preserved jsPlugins registration whose
// known external plugin package is no longer a project dependency.
// Preservation never drops entries, so a stale registration would keep
// failing oxlint at runtime until it is removed here. Hand-registered
// unknown packages are left alone — this tool cannot know their deps.
func warnOrphanedJsPlugins(cfg *config.OxlintConfig, detected []rule.ExternalPlugin) {
	installed := make(map[string]bool, len(detected))
	for _, p := range detected {
		installed[p.Package] = true
	}

	for _, pkg := range cfg.JsPlugins {
		if _, known := rule.ExternalPluginByPackage(pkg); !known || installed[pkg] {
			continue
		}

		slog.Warn("jsPlugins entry is not a project dependency; "+
			"remove it from the config or install the package", "package", pkg)
	}
}

// logExternalPluginHint tells the user that a detected external plugin was
// registered but its rules stay off, and where to enable them. Emitted only
// when a plugin was newly detected and no external rules exist yet, so an
// already-configured design system is not nagged.
func logExternalPluginHint(cfg *config.OxlintConfig, detected []rule.ExternalPlugin) {
	if len(detected) == 0 || config.HasExternalRules(cfg) {
		return
	}

	for _, p := range detected {
		slog.Info("registered external JS plugin without enabling rules; "+
			"enable its rules under \"rules\" with a \""+p.Prefix+"/\" prefix when ready",
			"plugin", p.Package,
			"docs", p.Docs,
		)
	}
}

func marshalConfigJSON(cfg *config.OxlintConfig) ([]byte, error) {
	data, err := cfg.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("generate config JSON: %w", err)
	}

	return data, nil
}

func writeConfig(cfg *config.OxlintConfig, targetPath string) error {
	data, err := marshalConfigJSON(cfg)
	if err != nil {
		return err
	}

	if err := atomicwrite.Write(targetPath, append(data, '\n')); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	slog.Info("configuration written", "path", targetPath)

	return nil
}

// writeDryRun prints cfg to stdout instead of writing a file. The shared
// marshal+check preamble with writeConfig is intentional idiomatic Go
// error propagation rather than duplicated logic — the actual output
// operations differ (file vs stdout).
func writeDryRun(cfg *config.OxlintConfig, targetPath string) error {
	data, err := marshalConfigJSON(cfg)
	if err != nil {
		return err
	}

	slog.Info("dry run", "path", targetPath)
	fmt.Println(string(data))

	return nil
}

func runFixIfNeeded(ctx context.Context, absRoot, targetPath string, runFix bool) error {
	if !runFix {
		return nil
	}

	slog.Info("running oxlint fix")

	fixResult, err := oxlint.RunFix(ctx, absRoot, targetPath)
	if err != nil {
		return fmt.Errorf("absRoot=%s: fix: %w", absRoot, err)
	}

	if fixResult.Output != "" {
		slog.Info("fix output", "detail", fixResult.Output)
	}

	slog.Info("fix complete")

	return nil
}

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
