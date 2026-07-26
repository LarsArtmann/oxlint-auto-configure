package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	finding "github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/pipeline"
	"github.com/larsartmann/oxlint-auto-configure/pkg/config"
	"github.com/larsartmann/oxlint-auto-configure/pkg/profile"
	"github.com/larsartmann/oxlint-auto-configure/pkg/rule"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildTestReport(t *testing.T) *finding.Report {
	t.Helper()

	f1 := finding.NewFinding(
		finding.RuleName("no-debugger"), finding.ToolName("oxlint"),
		"Unexpected debugger statement", finding.SeverityError,
		finding.Position{File: "test.ts", Line: 10, Column: 1}, 1.0)
	f1.Category = finding.CategoryCorrectness

	f2 := finding.NewFinding(
		finding.RuleName("no-explicit-any"), finding.ToolName("oxlint"),
		"Unexpected any", finding.SeverityWarning,
		finding.Position{File: "types.ts", Line: 5, Column: 3}, 1.0)
	f2.Category = "suspicious"

	report := finding.NewReport(finding.ToolInfo{Name: "oxlint", Version: "1.73.0"})
	report.AddFindings([]finding.Finding{f1, f2})
	report.ComputeSummary()

	return report
}

func buildTestPipelineResult() *pipeline.PipelineResult {
	return &pipeline.PipelineResult{
		TotalIterations: 1,
		Reason:          pipeline.ReasonStable,
	}
}

func TestRenderFindingsSummary(t *testing.T) {
	t.Parallel()

	require.NoError(t, renderFindings(FormatSummary, buildTestReport(t), buildTestPipelineResult(), ""))
}

func TestRenderFindingsJSON(t *testing.T) {
	t.Parallel()

	require.NoError(t, renderFindings(FormatJSON, buildTestReport(t), buildTestPipelineResult(), ""))
}

func TestRenderFindingsTable(t *testing.T) {
	t.Parallel()

	require.NoError(t, renderFindings(FormatTable, buildTestReport(t), buildTestPipelineResult(), ""))
}

func TestRenderFindingsSARIF(t *testing.T) {
	t.Parallel()

	require.NoError(t, renderFindings(FormatSARIF, buildTestReport(t), buildTestPipelineResult(), ""))
}

func TestRenderFindingsSARIFWithSeverity(t *testing.T) {
	t.Parallel()

	require.NoError(t, renderFindings(FormatSARIF, buildTestReport(t), buildTestPipelineResult(), finding.SeverityError))
}

func TestRenderFindingsReport(t *testing.T) {
	t.Parallel()

	require.NoError(t, renderFindings(FormatReport, buildTestReport(t), buildTestPipelineResult(), ""))
}

func TestRenderFindingsUnknownFormat(t *testing.T) {
	t.Parallel()

	err := renderFindings("invalid", buildTestReport(t), buildTestPipelineResult(), "")
	require.Error(t, err)
	assert.ErrorIs(t, err, errUnknownFormat)
}

func TestRenderFindingsSeverityFilterNoMatch(t *testing.T) {
	t.Parallel()

	require.NoError(t, renderFindings(FormatJSON, buildTestReport(t), buildTestPipelineResult(), finding.Severity("critical")))
}

func TestPrintSARIF(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	require.NoError(t, printSARIF(&buf, buildTestReport(t), ""))
	assert.NotEmpty(t, buf.String())
}

func TestPrintSARIFWithMinSeverity(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	require.NoError(t, printSARIF(&buf, buildTestReport(t), finding.SeverityError))
	assert.NotEmpty(t, buf.String())
}

func TestSortedByPosition(t *testing.T) {
	t.Parallel()

	findings := buildTestReport(t).ActiveFindings()
	sorted := sortedByPosition(findings)
	require.Len(t, sorted, len(findings))
}

func TestResolveConfigPath(t *testing.T) {
	t.Parallel()

	t.Run("default path when empty", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, filepath.Join("/project", defaultConfigPath), resolveConfigPath("", "/project"))
	})

	t.Run("uses provided path", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "/custom/path.json", resolveConfigPath("/custom/path.json", "/project"))
	})
}

func TestLogDiffIfExistingWithFile(t *testing.T) {
	t.Parallel()

	reg, err := rule.LoadRegistry()
	require.NoError(t, err)

	cfg, err := config.GenerateProjectConfig(profile.ProfileRecommended, reg, nil, nil)
	require.NoError(t, err)

	data, err := cfg.ToJSON()
	require.NoError(t, err)

	dir := t.TempDir()
	path := filepath.Join(dir, ".oxlintrc.json")
	require.NoError(t, os.WriteFile(path, data, 0o644))

	// Should not panic and should exercise the slog.Info path
	logDiffIfExisting(path, cfg)
}

func TestMarshalConfigJSON(t *testing.T) {
	t.Parallel()

	reg, err := rule.LoadRegistry()
	require.NoError(t, err)

	cfg, err := config.GenerateProjectConfig(profile.ProfileRecommended, reg, nil, nil)
	require.NoError(t, err)

	data, err := marshalConfigJSON(cfg)
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}
