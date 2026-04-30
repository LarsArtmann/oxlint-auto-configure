package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
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

func TestConfigureDryRunRecommended(t *testing.T) {
	t.Parallel()
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"configure", "--dry-run", "--root", t.TempDir()})

	err := cmd.Execute()
	require.NoError(t, err)
}

func TestConfigureDryRunMaximalTypesafe(t *testing.T) {
	t.Parallel()
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"configure", "--dry-run", "-p", "maximal-typesafe", "--root", t.TempDir()})

	err := cmd.Execute()
	require.NoError(t, err)
}

func TestConfigureInvalidProfile(t *testing.T) {
	t.Parallel()
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"configure", "-p", "invalid", "--root", t.TempDir()})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid profile")
}

func TestConfigureWritesFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".oxlintrc.json")

	cmd := NewRootCommand()
	cmd.SetArgs([]string{"configure", "-c", configPath, "--root", dir})

	err := cmd.Execute()
	require.NoError(t, err)

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	var cfg config.OxlintConfig
	err = json.Unmarshal(data, &cfg)
	require.NoError(t, err)
	assert.Equal(t, "error", cfg.Categories["correctness"])
}

func TestValidateConfig(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".oxlintrc.json")

	reg, err := rule.LoadRegistry()
	require.NoError(t, err)

	cat := profile.NewCategorizer(profile.ProfileRecommended, profile.PluginConfig{})
	gen := config.NewGenerator(cat, reg, nil)
	cfg := gen.Generate()
	data, err := cfg.ToJSON()
	require.NoError(t, err)
	err = os.WriteFile(configPath, append(data, '\n'), 0o644)
	require.NoError(t, err)

	cmd := NewRootCommand()
	cmd.SetArgs([]string{"validate", "-c", configPath})

	err = cmd.Execute()
	require.NoError(t, err)
}

func TestValidateUnknownRules(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".oxlintrc.json")

	badConfig := `{"rules":{"nonexistent-rule-xyz":"error"}}`
	err := os.WriteFile(configPath, []byte(badConfig), 0o644)
	require.NoError(t, err)

	cmd := NewRootCommand()
	cmd.SetArgs([]string{"validate", "-c", configPath})

	err = cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown rules")
}

func TestReportTable(t *testing.T) {
	t.Parallel()
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"report", "-f", "table", "--root", t.TempDir()})

	err := cmd.Execute()
	require.NoError(t, err)
}

func TestAnalyzeCleanProject(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"analyze", "--root", dir})

	err := cmd.Execute()
	require.NoError(t, err)
}

func TestVersionFlag(t *testing.T) {
	t.Parallel()
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"--version"})

	err := cmd.Execute()
	require.NoError(t, err)
}

func TestReportSummary(t *testing.T) {
	t.Parallel()
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"report", "-f", "summary", "--root", t.TempDir()})

	err := cmd.Execute()
	require.NoError(t, err)
}

func TestReportJSON(t *testing.T) {
	t.Parallel()
	reg, err := rule.LoadRegistry()
	require.NoError(t, err)

	cat := profile.NewCategorizer(profile.ProfileRecommended, profile.PluginConfig{})
	decisions := cat.DecideAll(reg)

	// Verify we can generate the report data
	assert.Len(t, decisions, 716)

	// Verify JSON serialization
	entries := make([]map[string]string, 0, len(decisions))
	for _, d := range decisions {
		entries = append(entries, map[string]string{
			"rule":     d.Rule.FullName(),
			"severity": string(d.Severity),
		})
	}
	data, err := json.Marshal(entries)
	require.NoError(t, err)
	assert.Greater(t, len(data), 1000)
}

func TestReportJSONDirect(t *testing.T) {
	t.Parallel()
	reg, err := rule.LoadRegistry()
	require.NoError(t, err)

	cat := profile.NewCategorizer(profile.ProfileRecommended, profile.PluginConfig{})
	decisions := cat.DecideAll(reg)

	var buf bytes.Buffer
	err = reportJSON(&buf, decisions)
	require.NoError(t, err)
	assert.NotEmpty(t, buf.String())

	var entries []map[string]any
	err = json.Unmarshal(buf.Bytes(), &entries)
	require.NoError(t, err)
	assert.Len(t, entries, 716)
}

func TestShowDiffExisting(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".oxlintrc.json")

	reg, err := rule.LoadRegistry()
	require.NoError(t, err)
	cat := profile.NewCategorizer(profile.ProfileRecommended, profile.PluginConfig{})
	gen := config.NewGenerator(cat, reg, nil)
	cfg := gen.Generate()
	data, err := cfg.ToJSON()
	require.NoError(t, err)
	err = os.WriteFile(configPath, append(data, '\n'), 0o600)
	require.NoError(t, err)

	diff := showDiffIfExisting(configPath, cfg)
	assert.NotEmpty(
		t,
		diff,
		"should produce diff summary when existing config matches (no changes)",
	)
}

func TestShowDiffMalformed(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".oxlintrc.json")
	err := os.WriteFile(configPath, []byte("not json"), 0o600)
	require.NoError(t, err)

	reg, err := rule.LoadRegistry()
	require.NoError(t, err)
	cat := profile.NewCategorizer(profile.ProfileRecommended, profile.PluginConfig{})
	gen := config.NewGenerator(cat, reg, nil)
	cfg := gen.Generate()

	diff := showDiffIfExisting(configPath, cfg)
	assert.Empty(t, diff, "malformed config should return empty diff")
}

func TestShowDiffMissing(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".oxlintrc.json")

	reg, err := rule.LoadRegistry()
	require.NoError(t, err)
	cat := profile.NewCategorizer(profile.ProfileRecommended, profile.PluginConfig{})
	gen := config.NewGenerator(cat, reg, nil)
	cfg := gen.Generate()

	diff := showDiffIfExisting(configPath, cfg)
	assert.Empty(t, diff, "missing file should return empty diff")
}

func TestValidateInvalidSeverity(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".oxlintrc.json")

	badConfig := `{"rules":{"no-debugger":"badsev"}}`
	err := os.WriteFile(configPath, []byte(badConfig), 0o600)
	require.NoError(t, err)

	err = Validate(configPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid severities")
}

func TestSetupLoggingVerboseQuietConflict(t *testing.T) {
	t.Parallel()
	err := setupLogging(true, true, io.Discard)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot use both")
}

func TestSetupLoggingVerbose(t *testing.T) {
	t.Parallel()
	err := setupLogging(true, false, io.Discard)
	require.NoError(t, err)
}

func TestSetupLoggingQuiet(t *testing.T) {
	t.Parallel()
	err := setupLogging(false, true, io.Discard)
	require.NoError(t, err)
}

func TestCompactLogAttrStripsTime(t *testing.T) {
	t.Parallel()
	attr := compactLogAttr(nil, slog.Attr{Key: slog.TimeKey})
	assert.Equal(t, slog.Attr{}, attr)
}

func TestCompactLogAttrPreservesOther(t *testing.T) {
	t.Parallel()
	original := slog.Attr{Key: slog.MessageKey, Value: slog.StringValue("test")}
	attr := compactLogAttr(nil, original)
	assert.Equal(t, original, attr)
}

func TestProfileNames(t *testing.T) {
	t.Parallel()
	names := profile.AllProfileNames()
	assert.Len(t, names, 4)
	assert.Contains(t, names, "recommended")
}

func TestPrintFormatErrorNil(t *testing.T) {
	t.Parallel()
	err := printFormatError(nil, "json")
	assert.NoError(t, err)
}

func TestPrintFormatErrorWithErr(t *testing.T) {
	t.Parallel()
	err := printFormatError(errors.New("write failed"), "table")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "print table")
}

func TestSummaryFromReport(t *testing.T) {
	t.Parallel()
	report := finding.NewReport(finding.ToolInfo{Name: "test", Version: "0.0.0"})
	f := finding.NewFinding("rule1", "test", "msg", finding.SeverityError,
		finding.Position{File: "a.ts", Line: 1})
	f.Category = finding.CategoryCorrectness
	report.AddFindings([]finding.Finding{f})
	report.ComputeSummary()

	sv := summaryFromReport(report, &pipeline.PipelineResult{
		TotalIterations: 2,
		Stable:          true,
	})
	assert.Equal(t, 1, sv.Total)
	assert.True(t, sv.Stable)
	assert.Equal(t, 2, sv.Iterations)
}

func TestFindingsToViews(t *testing.T) {
	t.Parallel()
	findings := []finding.Finding{
		finding.NewFinding("no-debugger", "oxlint", "msg",
			finding.SeverityWarning,
			finding.Position{File: "test.ts", Line: 5, Column: 3}),
	}
	views := findingsToViews(findings)
	require.Len(t, views, 1)
	assert.Equal(t, "no-debugger", views[0].Rule)
	assert.Equal(t, "test.ts", views[0].File)
	assert.Equal(t, 5, views[0].Line)
}

func TestPrintReportJSON(t *testing.T) {
	t.Parallel()
	report := finding.NewReport(finding.ToolInfo{Name: "oxlint", Version: "1.0.0"})
	f := finding.NewFinding("no-unused-vars", "oxlint", "unused variable",
		finding.SeverityError,
		finding.Position{File: "a.ts", Line: 10, Column: 5})
	f.Category = finding.CategoryCorrectness
	f.FixStrategy = finding.FixStrategyDirect
	report.AddFindings([]finding.Finding{f})
	report.ComputeSummary()

	var buf bytes.Buffer
	err := printReportJSON(&buf, report)
	require.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(buf.Bytes(), &parsed)
	require.NoError(t, err)

	tool := parsed["tool"].(map[string]any)
	assert.Equal(t, "oxlint", tool["name"])
	assert.Equal(t, "1.0.0", tool["version"])

	summary := parsed["summary"].(map[string]any)
	assert.Equal(t, float64(1), summary["total"])

	findings := parsed["findings"].([]any)
	assert.Len(t, findings, 1)
}

func TestParseOptionalSeverity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  finding.Severity
		err   bool
	}{
		{"", "", false},
		{"error", finding.SeverityError, false},
		{"warning", finding.SeverityWarning, false},
		{"info", finding.SeverityInfo, false},
		{"bad", "", true},
	}

	for _, tt := range tests {
		got, err := parseOptionalSeverity(tt.input)
		if tt.err {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		}
	}
}

func TestActiveWithFilter(t *testing.T) {
	t.Parallel()

	report := finding.NewReport(finding.ToolInfo{Name: "test"})
	report.AddFindings([]finding.Finding{
		finding.NewFinding("r1", "test", "msg", finding.SeverityError,
			finding.Position{File: "a.ts", Line: 1}),
		finding.NewFinding("r2", "test", "msg", finding.SeverityWarning,
			finding.Position{File: "b.ts", Line: 2}),
		finding.NewFinding("r3", "test", "msg", finding.SeverityInfo,
			finding.Position{File: "c.ts", Line: 3}),
	})
	report.ComputeSummary()

	all := activeWithFilter(report, "")
	assert.Len(t, all, 3)

	errors := activeWithFilter(report, finding.SeverityError)
	assert.Len(t, errors, 1)
	assert.Equal(t, "r1", errors[0].Rule)

	warnings := activeWithFilter(report, finding.SeverityWarning)
	assert.Len(t, warnings, 2)
}
