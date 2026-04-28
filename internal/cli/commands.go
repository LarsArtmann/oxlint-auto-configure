// Package cli implements the oxlint-auto-configure CLI commands.
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	finding "github.com/larsartmann/go-finding"
	"github.com/larsartmann/oxlint-auto-configure/pkg/config"
	"github.com/larsartmann/oxlint-auto-configure/pkg/detect"
	"github.com/larsartmann/oxlint-auto-configure/pkg/diff"
	"github.com/larsartmann/oxlint-auto-configure/pkg/oxlint"
	"github.com/larsartmann/oxlint-auto-configure/pkg/profile"
	"github.com/larsartmann/oxlint-auto-configure/pkg/rule"
	"github.com/spf13/cobra"
)

const (
	defaultConfigPath = ".oxlintrc.json"
	defaultProfile   = profile.ProfileRecommended
)

var version = "dev"

// NewRootCommand creates the root CLI command.
func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "oxlint-auto-configure",
		Short: "Automatically configure oxlint for maximum type safety",
		Long: `oxlint-auto-configure analyzes your project and generates the optimal
.oxlintrc.json configuration for maximum type safety and correctness enforcement.

It uses the go-finding library to run oxlint, collect findings, and
auto-configure every available rule with the best severity setting.`,
		Version: version,
	}

	root.AddCommand(newConfigureCommand())
	root.AddCommand(newAnalyzeCommand())
	root.AddCommand(newValidateCommand())
	root.AddCommand(newReportCommand())

	return root
}

func newConfigureCommand() *cobra.Command {
	var (
		profileFlag string
		configPath  string
		dryRun      bool
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
				return fmt.Errorf("invalid profile %q: choose from %s", profileFlag, strings.Join(profileNames(), ", "))
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

			fmt.Fprintf(os.Stderr, "Detected project: %s\n", detect.FormatTypes(projectTypes))
			fmt.Fprintf(os.Stderr, "Profile: %s\n", p)
			fmt.Fprintf(os.Stderr, "Total rules: %d\n", reg.Len())

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

			// Show diff if existing config exists
			if existingData, err := os.ReadFile(targetPath); err == nil {
				existing, err := config.FromJSON(existingData)
				if err == nil {
					d := diff.NewDiffer(existing, cfg)
					fmt.Fprintf(os.Stderr, "\nChanges:\n%s\n", d.FormatDiff())
					fmt.Fprintf(os.Stderr, "%s\n", d.Summary())
				}
			}

			data, err := cfg.ToJSON()
			if err != nil {
				return fmt.Errorf("generate config JSON: %w", err)
			}

			if err := os.WriteFile(targetPath, append(data, '\n'), 0o644); err != nil {
				return fmt.Errorf("write config: %w", err)
			}

			fmt.Fprintf(os.Stderr, "Configuration written to %s\n", targetPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&profileFlag, "profile", "p", string(defaultProfile), "Configuration profile")
	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Output config file path (default: .oxlintrc.json)")
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Show what would change without writing")
	cmd.Flags().StringVar(&rootDir, "root", ".", "Project root directory")

	return cmd
}

func newAnalyzeCommand() *cobra.Command {
	var rootDir string

	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze current project and show findings using go-finding pipeline",
		RunE: func(cmd *cobra.Command, args []string) error {
			if rootDir == "" {
				rootDir = "."
			}

			absRoot, err := filepath.Abs(rootDir)
			if err != nil {
				return fmt.Errorf("resolve root dir: %w", err)
			}

			reg, err := rule.LoadRegistry()
			if err != nil {
				return fmt.Errorf("load rule registry: %w", err)
			}

			fmt.Fprintf(os.Stderr, "Loaded %d rules from registry\n", reg.Len())

			// Run oxlint via go-finding detector
			configPath := filepath.Join(absRoot, defaultConfigPath)
			var opts []oxlint.Option
			if _, err := os.Stat(configPath); err == nil {
				opts = append(opts, oxlint.WithConfig(configPath))
			}

			detector := oxlint.NewDetector(absRoot, opts...)
			findings, err := detector.Detect(cmd.Context())
			if err != nil {
				return fmt.Errorf("run oxlint: %w", err)
			}

			if len(findings) == 0 {
				fmt.Fprintln(os.Stderr, "No findings — your project is clean!")
				return nil
			}

			report := finding.NewReport(finding.ToolInfo{Name: "oxlint", Version: version})
			report.AddFindings(findings)
			report.ComputeSummary()

			fmt.Fprintf(os.Stderr, "\nFindings: %d total\n", report.Summary.Total)
			fmt.Fprintf(os.Stderr, "  By severity: %v\n", report.Summary.BySeverity)
			fmt.Fprintf(os.Stderr, "  By category: %v\n", report.Summary.ByCategory)
			fmt.Fprintf(os.Stderr, "  Files affected: %d\n", report.Summary.FilesAffected)

			sarif, err := report.ToSARIF()
			if err != nil {
				return fmt.Errorf("generate SARIF: %w", err)
			}
			fmt.Println(string(sarif))

			return nil
		},
	}

	cmd.Flags().StringVar(&rootDir, "root", ".", "Project root directory")

	return cmd
}

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

			// Validate that all rules in the config are known
			var unknown []string
			for name := range cfg.Rules {
				if _, ok := reg.ByName(name); !ok {
					unknown = append(unknown, name)
				}
			}

			if len(unknown) > 0 {
				fmt.Fprintf(os.Stderr, "Unknown rules: %s\n", strings.Join(unknown, ", "))
				return fmt.Errorf("%d unknown rules found", len(unknown))
			}

			// Validate severity values
			validSeverities := map[string]bool{"error": true, "warn": true, "off": true}
			var invalid []string
			for name, sev := range cfg.Rules {
				if !validSeverities[sev] {
					invalid = append(invalid, fmt.Sprintf("%s=%s", name, sev))
				}
			}

			if len(invalid) > 0 {
				fmt.Fprintf(os.Stderr, "Invalid severities: %s\n", strings.Join(invalid, ", "))
				return fmt.Errorf("%d invalid severities found", len(invalid))
			}

			enabled := 0
			for _, sev := range cfg.Rules {
				if sev != "off" {
					enabled++
				}
			}

			fmt.Fprintf(os.Stderr, "Config valid: %s\n", targetPath)
			fmt.Fprintf(os.Stderr, "  Rules: %d total, %d enabled, %d disabled\n",
				len(cfg.Rules), enabled, len(cfg.Rules)-enabled)
			fmt.Fprintf(os.Stderr, "  Plugins: %v\n", cfg.Plugins)
			fmt.Fprintf(os.Stderr, "  Categories: %v\n", cfg.Categories)

			return nil
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Config file path (default: .oxlintrc.json)")

	return cmd
}

func newReportCommand() *cobra.Command {
	var (
		profileFlag string
		format      string
		rootDir     string
	)

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate a report of all rules and their recommended severity",
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
				return fmt.Errorf("invalid profile %q", profileFlag)
			}

			reg, err := rule.LoadRegistry()
			if err != nil {
				return fmt.Errorf("load registry: %w", err)
			}

			det := detect.NewDetector(absRoot)
			pluginConfig, _, err := det.Detect()
			if err != nil {
				return fmt.Errorf("detect project type: %w", err)
			}

			cat := profile.NewCategorizer(p, pluginConfig)
			decisions := cat.DecideAll(reg)

			switch format {
			case "json":
				return reportJSON(decisions)
			case "summary":
				return reportSummary(decisions, reg)
			default:
				return reportTable(decisions)
			}
		},
	}

	cmd.Flags().StringVarP(&profileFlag, "profile", "p", string(defaultProfile), "Configuration profile")
	cmd.Flags().StringVarP(&format, "format", "f", "table", "Output format: table, json, summary")
	cmd.Flags().StringVar(&rootDir, "root", ".", "Project root directory")

	return cmd
}

func writeDryRun(cfg *config.OxlintConfig, targetPath string) error {
	data, err := cfg.ToJSON()
	if err != nil {
		return fmt.Errorf("generate config JSON: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Dry run — would write to %s:\n", targetPath)
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

func reportJSON(decisions []profile.RuleDecision) error {
	type entry struct {
		Rule     string `json:"rule"`
		Plugin   string `json:"plugin"`
		Category string `json:"category"`
		Severity string `json:"severity"`
		Default  bool   `json:"default"`
		Fixable  bool   `json:"fixable"`
	}

	entries := make([]entry, 0, len(decisions))
	for _, d := range decisions {
		entries = append(entries, entry{
			Rule:     d.Rule.FullName(),
			Plugin:   string(d.Rule.Plugin),
			Category: string(d.Rule.Category),
			Severity: string(d.Severity),
			Default:  d.Rule.Enabled,
			Fixable:  d.Rule.IsFixable(),
		})
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func reportSummary(decisions []profile.RuleDecision, reg *rule.Registry) error {
	counts := make(map[string]int)
	for _, d := range decisions {
		counts[string(d.Severity)]++
	}

	fmt.Fprintf(os.Stderr, "Total rules: %d\n", reg.Len())
	fmt.Fprintf(os.Stderr, "  error: %d\n", counts["error"])
	fmt.Fprintf(os.Stderr, "  warn:  %d\n", counts["warn"])
	fmt.Fprintf(os.Stderr, "  off:   %d\n", counts["off"])
	fmt.Fprintf(os.Stderr, "  Enabled by default: %d\n", len(reg.EnabledByDefault()))
	fmt.Fprintf(os.Stderr, "  Disabled by default: %d\n", len(reg.DisabledByDefault()))
	fmt.Fprintf(os.Stderr, "  Fixable: %d\n", len(reg.Fixable()))
	fmt.Fprintf(os.Stderr, "  Type-aware: %d\n", len(reg.TypeAwareRules()))

	return nil
}

func reportTable(decisions []profile.RuleDecision) error {
	fmt.Println("| Rule | Plugin | Category | Default | Severity |")
	fmt.Println("|------|--------|----------|---------|----------|")

	for _, d := range decisions {
		def := "off"
		if d.Rule.Enabled {
			def = "on"
		}
		fmt.Printf("| %s | %s | %s | %s | %s |\n",
			d.Rule.FullName(), d.Rule.Plugin, d.Rule.Category, def, d.Severity)
	}

	return nil
}
