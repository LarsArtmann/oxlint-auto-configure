// Package cli implements the oxlint-auto-configure CLI commands.
package cli

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	finding "github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/pipeline"
	"github.com/larsartmann/oxlint-auto-configure/pkg/format"
	"github.com/larsartmann/oxlint-auto-configure/pkg/oxlint"
	"github.com/larsartmann/oxlint-auto-configure/pkg/rule"
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
  report   Full go-finding Report JSON (tool info, summary, all fields)
  sarif    SARIF format to stdout (for CI/GitHub integration)
  table    Markdown table to stdout`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runAnalyze(cmd.Context(), rootDir, formatFlag)
		},
	}

	cmd.Flags().StringVar(&rootDir, "root", ".", "Project root directory")
	cmd.Flags().
		StringVarP(&formatFlag, "format", "f", "summary", "Output format: summary, json, report, sarif, table")

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

	reg, err := rule.LoadRegistry()
	if err != nil {
		return fmt.Errorf("load rules: %w", err)
	}

	oxlintVersion, err := oxlint.CheckVersion(ctx)
	if err != nil {
		slog.Warn("could not determine oxlint version", "error", err)
	}

	configPath := filepath.Join(absRoot, defaultConfigPath)
	var opts []oxlint.Option
	if _, err := os.Stat(configPath); err == nil {
		opts = append(opts, oxlint.WithConfig(configPath))
	}
	opts = append(opts, oxlint.WithRegistry(reg))

	detector := oxlint.NewDetector(absRoot, opts...)

	metrics := pipeline.NewMetrics()
	pipelineCfg := pipeline.DefaultConfig()
	pipelineCfg.DryRun = true
	pipelineCfg.VerifyAfterFix = false
	pipelineCfg.GracefulDegradation = true
	pipelineCfg.Metrics = metrics
	pipelineCfg.Retry = &pipeline.RetryConfig{
		MaxRetries: 2,
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   2 * time.Second,
	}
	pipelineCfg.OnFinding = func(f finding.Finding) {
		slog.Debug("finding", "rule", f.Rule, "file", f.Position.File, "line", f.Position.Line)
	}
	pipelineCfg.OnIteration = func(iter int, findings []finding.Finding) {
		slog.Info("iteration complete", "iter", iter, "findings", len(findings))
	}

	p, err := pipeline.New(pipelineCfg, absRoot, detector)
	if err != nil {
		return fmt.Errorf("create pipeline: %w", err)
	}

	result, err := p.Run(ctx)
	if err != nil {
		return fmt.Errorf("pipeline: %w", err)
	}

	if snap := result.Metrics; !snap.StartTime.IsZero() {
		slog.Info("pipeline metrics",
			"duration", snap.TotalDuration.String(),
			"fixes_applied", snap.FixesApplied,
		)
	}

	if result.TotalDetected == 0 {
		slog.Info("no findings — project is clean")

		return nil
	}

	report := finding.NewReport(finding.ToolInfo{Name: "oxlint", Version: oxlintVersion})
	for _, iter := range result.Iterations {
		report.AddFindings(iter.Findings())
	}
	report.ComputeSummary()

	return renderFindings(formatFlag, report, result)
}

// renderFindings converts go-finding types to format views and delegates rendering.
func renderFindings(fmtFlag string, report *finding.Report, result *pipeline.PipelineResult) error {
	sv := summaryFromReport(report, result)

	switch fmtFlag {
	case "summary":
		return printFormatError(format.PrintSummary(os.Stderr, sv), "summary")
	case "json":
		views := findingsToViews(report.ActiveFindings())
		return printFormatError(format.PrintFindingsJSON(os.Stdout, views), "json")
	case "report":
		return printReportJSON(os.Stdout, report)
	case "table":
		sorted := report.ActiveFindings()
		finding.SortByPosition(sorted)
		views := findingsToViews(sorted)
		return printFormatError(format.PrintFindingsTable(os.Stdout, views), "table")
	case "sarif":
		return printSARIF(os.Stdout, report)
	default:
		return fmt.Errorf("unknown format %q: choose from summary, json, report, sarif, table", fmtFlag)
	}
}

func printFormatError(err error, label string) error {
	if err != nil {
		return fmt.Errorf("print %s: %w", label, err)
	}
	return nil
}

// findingsToViews converts go-finding Finding values to format views.
func findingsToViews(findings []finding.Finding) []format.FindingView {
	views := make([]format.FindingView, 0, len(findings))
	for _, f := range findings {
		views = append(views, format.FindingView{
			Rule:        f.Rule,
			Message:     f.Message,
			Severity:    string(f.Severity),
			Category:    string(f.Category),
			File:        f.Position.File,
			Line:        f.Position.Line,
			Column:      f.Position.Column,
			DocsURL:     f.Metadata["url"],
			FixStrategy: string(f.FixStrategy),
			Tag:         f.Tag,
			Snippet:     f.Snippet,
		})
	}
	return views
}

// summaryFromReport builds a SummaryView from a go-finding Report and pipeline result.
func summaryFromReport(
	report *finding.Report,
	result *pipeline.PipelineResult,
) *format.SummaryView {
	bySev := make(map[string]int, len(report.Summary.BySeverity))
	for k, v := range report.Summary.BySeverity {
		bySev[string(k)] = v
	}
	byCat := make(map[string]int, len(report.Summary.ByCategory))
	for k, v := range report.Summary.ByCategory {
		byCat[string(k)] = v
	}

	return &format.SummaryView{
		Total:         report.Summary.Total,
		BySeverity:    bySev,
		ByCategory:    byCat,
		FilesAffected: report.Summary.FilesAffected,
		Iterations:    result.TotalIterations,
		Stable:        result.Stable,
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

// printReportJSON renders the full go-finding Report as JSON using the library's
// native serialization (includes tool info, summary, and all finding fields).
func printReportJSON(w io.Writer, report *finding.Report) error {
	json, err := report.PrettyJSON()
	if err != nil {
		return fmt.Errorf("serialize report: %w", err)
	}
	_, _ = fmt.Fprintln(w, json)

	return nil
}
