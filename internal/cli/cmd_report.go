package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/larsartmann/oxlint-auto-configure/pkg/config"
	"github.com/larsartmann/oxlint-auto-configure/pkg/detect"
	"github.com/larsartmann/oxlint-auto-configure/pkg/profile"
	"github.com/larsartmann/oxlint-auto-configure/pkg/rule"
	"github.com/spf13/cobra"
)

func newReportCommand() *cobra.Command {
	var (
		format  string
		rootDir string
	)

	cmd := &cobra.Command{
		Use:   CmdReport,
		Short: "Generate a report of all rules and their recommended severity",
		RunE: func(_ *cobra.Command, _ []string) error {
			if rootDir == "" {
				rootDir = "."
			}

			absRoot, err := filepath.Abs(rootDir)
			if err != nil {
				return fmt.Errorf("resolve root dir: %w", err)
			}

			p := profile.Profile(profileFlag)
			if !p.IsValid() {
				return fmt.Errorf("%w %q", config.ErrInvalidProfile, profileFlag)
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
			case FormatJSON:
				return reportJSON(os.Stdout, decisions)
			case FormatSummary:
				return reportSummary(decisions, reg)
			default:
				return reportTable(decisions)
			}
		},
	}

	AddProfileFlag(cmd)
	cmd.Flags().StringVarP(&format, "format", "f", "table", "Output format: table, json, summary")
	cmd.Flags().StringVar(&rootDir, "root", ".", "Project root directory")

	return cmd
}

func reportJSON(w io.Writer, decisions []profile.RuleDecision) error {
	type entry struct {
		Rule      string `json:"rule"`
		Plugin    string `json:"plugin"`
		Category  string `json:"category"`
		Severity  string `json:"severity"`
		Default   bool   `json:"default"`
		Fixable   bool   `json:"fixable"`
		TypeAware bool   `json:"type_aware"`
		DocsURL   string `json:"docs_url,omitempty"`
	}

	entries := make([]entry, 0, len(decisions))
	for _, d := range decisions {
		entries = append(entries, entry{
			Rule:      d.Rule.FullName(),
			Plugin:    string(d.Rule.Plugin),
			Category:  string(d.Rule.Category),
			Severity:  string(d.Severity),
			Default:   d.Rule.Enabled,
			Fixable:   d.Rule.IsFixable(),
			TypeAware: d.Rule.TypeAware,
			DocsURL:   d.Rule.DocsURL,
		})
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal report: %w", err)
	}
	_, _ = fmt.Fprintln(w, string(data))

	return nil
}

func reportSummary(decisions []profile.RuleDecision, reg *rule.Registry) error {
	counts := make(map[string]int)
	for _, d := range decisions {
		counts[string(d.Severity)]++
	}

	slog.Info(
		"rule summary",
		"total", reg.Len(),
		"error", counts["error"],
		"warn", counts["warn"],
		"off", counts["off"],
		"enabled_by_default", len(reg.EnabledByDefault()),
		"disabled_by_default", len(reg.DisabledByDefault()),
		"fixable", len(reg.Fixable()),
		"type_aware", len(reg.TypeAwareRules()),
	)

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
