// Package cli implements the oxlint-auto-configure CLI commands.
package cli

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	finding "github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/pipeline"
	"github.com/larsartmann/oxlint-auto-configure/pkg/format"
	"github.com/larsartmann/oxlint-auto-configure/pkg/oxlint"
	"github.com/spf13/cobra"
)

func newAnalyzeCommand() *cobra.Command {
	var (
		rootDir    string
		formatFlag string
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runAnalyze(cmd.Context(), rootDir, formatFlag)
		},
	}

	cmd.Flags().StringVar(&rootDir, "root", ".", "Project root directory")
	cmd.Flags().
		StringVarP(&formatFlag, "format", "f", "summary", "Output format: summary, json, sarif, table")

	return cmd
}

func runAnalyze(ctx context.Context, rootDir, formatFlag string) error {
	if rootDir == "" {
		rootDir = "."
	}

	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return fmt.Errorf("resolve root dir: %w", err)
	}

	if err := oxlint.CheckBinary(ctx); err != nil {
		return fmt.Errorf("check oxlint: %w", err)
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

	result, err := p.Run(ctx)
	if err != nil {
		return fmt.Errorf("pipeline: %w", err)
	}

	if result.TotalDetected == 0 {
		slog.Info("no findings — project is clean")

		return nil
	}

	report := finding.NewReport(finding.ToolInfo{Name: "oxlint", Version: version})
	for _, iter := range result.Iterations {
		report.AddFindings(iter.Findings())
	}
	report.ComputeSummary()

	return renderFindings(formatFlag, report, result)
}

// renderFindings converts go-finding types to format views and delegates rendering.
func renderFindings(fmtFlag string, report *finding.Report, result *pipeline.PipelineResult) error {
	views := make([]format.FindingView, 0, len(report.Findings))
	for _, f := range report.Findings {
		views = append(views, format.FindingView{
			Rule:     f.Rule,
			Message:  f.Message,
			Severity: string(f.Severity),
			Category: string(f.Category),
			File:     f.Position.File,
			Line:     f.Position.Line,
			Column:   f.Position.Column,
		})
	}

	bySev := make(map[string]int, len(report.Summary.BySeverity))
	for k, v := range report.Summary.BySeverity {
		bySev[string(k)] = v
	}
	byCat := make(map[string]int, len(report.Summary.ByCategory))
	for k, v := range report.Summary.ByCategory {
		byCat[string(k)] = v
	}

	sv := &format.SummaryView{
		Total:         report.Summary.Total,
		BySeverity:    bySev,
		ByCategory:    byCat,
		FilesAffected: report.Summary.FilesAffected,
		Iterations:    result.TotalIterations,
		Stable:        result.Stable,
	}

	switch fmtFlag {
	case "summary":
		if err := format.PrintSummary(os.Stderr, sv); err != nil {
			return fmt.Errorf("print summary: %w", err)
		}
		return nil
	case "json":
		if err := format.PrintFindingsJSON(os.Stdout, views); err != nil {
			return fmt.Errorf("print json: %w", err)
		}
		return nil
	case "table":
		if err := format.PrintFindingsTable(os.Stdout, views); err != nil {
			return fmt.Errorf("print table: %w", err)
		}
		return nil
	case "sarif":
		return printSARIF(os.Stdout, report)
	default:
		return fmt.Errorf("unknown format %q: choose from summary, json, sarif, table", fmtFlag)
	}
}

// printSARIF renders SARIF directly from go-finding Report (requires go-finding method).
func printSARIF(w io.Writer, report *finding.Report) error {
	sarif, err := report.ToSARIF()
	if err != nil {
		return fmt.Errorf("generate SARIF: %w", err)
	}
	_, _ = fmt.Fprintln(w, string(sarif))

	return nil
}
