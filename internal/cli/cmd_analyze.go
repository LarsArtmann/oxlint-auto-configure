package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	finding "github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/pipeline"
	"github.com/larsartmann/oxlint-auto-configure/pkg/oxlint"
	"github.com/spf13/cobra"
)

func newAnalyzeCommand() *cobra.Command {
	var (
		rootDir string
		format  string
	)

	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze current project using go-finding pipeline (detect → triage → report)",
		Long: `Run oxlint through the go-finding pipeline and output findings.

Formats:
  summary  Human-readable summary to stderr (default)
  json     JSON array of findings to stdout
  sarif    SARIF format to stdout (for CI/GitHub integration)
  table    Markdown table to stdout`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if rootDir == "" {
				rootDir = "."
			}

			absRoot, err := filepath.Abs(rootDir)
			if err != nil {
				return fmt.Errorf("resolve root dir: %w", err)
			}

			if err := oxlint.CheckBinary(cmd.Context()); err != nil {
				return err
			}

			configPath := filepath.Join(absRoot, defaultConfigPath)
			var opts []oxlint.Option
			if _, err := os.Stat(configPath); err == nil {
				opts = append(opts, oxlint.WithConfig(configPath))
			}

			detector := oxlint.NewDetector(absRoot, opts...)
			pipelineCfg := pipeline.DefaultConfig()
			pipelineCfg.DryRun = true
			pipelineCfg.VerifyAfterFix = false
			pipelineCfg.GracefulDegradation = true

			p, err := pipeline.New(pipelineCfg, absRoot, detector)
			if err != nil {
				return fmt.Errorf("create pipeline: %w", err)
			}

			result, err := p.Run(cmd.Context())
			if err != nil {
				return fmt.Errorf("pipeline: %w", err)
			}

			if result.TotalDetected == 0 {
				fmt.Fprintln(os.Stderr, "No findings — your project is clean!")
				return nil
			}

			report := finding.NewReport(finding.ToolInfo{Name: "oxlint", Version: version})
			for _, iter := range result.Iterations {
				report.AddFindings(iter.Findings())
			}
			report.ComputeSummary()

			switch format {
			case "sarif":
				return printSARIF(report)
			case "json":
				return printFindingsJSON(report)
			case "table":
				return printFindingsTable(report)
			case "summary":
				return printSummary(report, result)
			default:
				return fmt.Errorf("unknown format %q: choose from summary, json, sarif, table", format)
			}
		},
	}

	cmd.Flags().StringVar(&rootDir, "root", ".", "Project root directory")
	cmd.Flags().StringVarP(&format, "format", "f", "summary", "Output format: summary, json, sarif, table")

	return cmd
}

func printSummary(report *finding.Report, result *pipeline.PipelineResult) error {
	fmt.Fprintf(os.Stderr, "Findings: %d total\n", report.Summary.Total)
	fmt.Fprintf(os.Stderr, "  By severity: %v\n", formatSeverityMap(report.Summary.BySeverity))
	fmt.Fprintf(os.Stderr, "  By category: %v\n", formatCategoryMap(report.Summary.ByCategory))
	fmt.Fprintf(os.Stderr, "  Files affected: %d\n", report.Summary.FilesAffected)
	fmt.Fprintf(os.Stderr, "  Pipeline iterations: %d, stable: %v\n", result.TotalIterations, result.Stable)
	return nil
}

func printSARIF(report *finding.Report) error {
	sarif, err := report.ToSARIF()
	if err != nil {
		return fmt.Errorf("generate SARIF: %w", err)
	}
	fmt.Println(string(sarif))
	return nil
}

func printFindingsJSON(report *finding.Report) error {
	type entry struct {
		Rule     string `json:"rule"`
		Message  string `json:"message"`
		Severity string `json:"severity"`
		Category string `json:"category"`
		File     string `json:"file"`
		Line     int    `json:"line"`
		Column   int    `json:"column"`
	}

	entries := make([]entry, 0, len(report.Findings))
	for _, f := range report.Findings {
		entries = append(entries, entry{
			Rule:     f.Rule,
			Message:  f.Message,
			Severity: string(f.Severity),
			Category: string(f.Category),
			File:     f.Position.File,
			Line:     f.Position.Line,
			Column:   f.Position.Column,
		})
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal findings: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func printFindingsTable(report *finding.Report) error {
	fmt.Println("| Rule | Severity | Category | File:Line | Message |")
	fmt.Println("|------|----------|----------|-----------|---------|")

	sorted := make([]finding.Finding, len(report.Findings))
	copy(sorted, report.Findings)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Position.File != sorted[j].Position.File {
			return sorted[i].Position.File < sorted[j].Position.File
		}
		return sorted[i].Position.Line < sorted[j].Position.Line
	})

	for _, f := range sorted {
		loc := fmt.Sprintf("%s:%d", f.Position.File, f.Position.Line)
		msg := f.Message
		if len(msg) > 60 {
			msg = msg[:57] + "..."
		}
		msg = strings.ReplaceAll(msg, "|", "\\|")
		fmt.Printf("| %s | %s | %s | %s | %s |\n",
			f.Rule, f.Severity, f.Category, loc, msg)
	}
	return nil
}

func formatSeverityMap(m map[finding.Severity]int) string {
	parts := make([]string, 0, len(m))
	for k, v := range m {
		parts = append(parts, fmt.Sprintf("%s=%d", k, v))
	}
	sort.Strings(parts)
	return strings.Join(parts, ", ")
}

func formatCategoryMap(m map[finding.Category]int) string {
	parts := make([]string, 0, len(m))
	for k, v := range m {
		parts = append(parts, fmt.Sprintf("%s=%d", k, v))
	}
	sort.Strings(parts)
	return strings.Join(parts, ", ")
}
