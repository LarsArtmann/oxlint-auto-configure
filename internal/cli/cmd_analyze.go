package cli

import (
	"fmt"
	"os"
	"path/filepath"

	finding "github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/pipeline"
	"github.com/larsartmann/oxlint-auto-configure/pkg/oxlint"
	"github.com/spf13/cobra"
)

func newAnalyzeCommand() *cobra.Command {
	var rootDir string

	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze current project using go-finding pipeline (detect → triage → report)",
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

			fmt.Fprintf(os.Stderr, "\nFindings: %d total\n", report.Summary.Total)
			fmt.Fprintf(os.Stderr, "  By severity: %v\n", report.Summary.BySeverity)
			fmt.Fprintf(os.Stderr, "  By category: %v\n", report.Summary.ByCategory)
			fmt.Fprintf(os.Stderr, "  Files affected: %d\n", report.Summary.FilesAffected)
			fmt.Fprintf(
				os.Stderr,
				"  Pipeline iterations: %d, stable: %v\n",
				result.TotalIterations,
				result.Stable,
			)

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
