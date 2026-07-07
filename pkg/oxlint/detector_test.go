package oxlint

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	finding "github.com/larsartmann/go-finding"
	"github.com/larsartmann/oxlint-auto-configure/pkg/rule"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const realOxlintOutput = `{
  "diagnostics": [
    {
      "message": "` + "`" + `debugger` + "`" + ` statement is not allowed",
      "code": "eslint(no-debugger)",
      "severity": "warning",
      "causes": [],
      "url": "https://oxc.rs/docs/guide/usage/linter/rules/eslint/no-debugger.html",
      "help": "Remove the debugger statement",
      "filename": "test.ts",
      "labels": [
        {
          "span": {
            "offset": 18,
            "length": 9,
            "line": 2,
            "column": 1
          }
        }
      ],
      "related": []
    },
    {
      "message": "Variable 'x' is declared but never used.",
      "code": "eslint(no-unused-vars)",
      "severity": "warning",
      "causes": [],
      "url": "https://oxc.rs/docs/guide/usage/linter/rules/eslint/no-unused-vars.html",
      "help": "Consider removing this declaration.",
      "filename": "src/main.ts",
      "labels": [
        {
          "label": "'x' is declared here",
          "span": {
            "offset": 6,
            "length": 1,
            "line": 1,
            "column": 7
          }
        }
      ],
      "related": []
    },
    {
      "message": "Unexpected any. Specify a different type.",
      "code": "typescript/no-explicit-any",
      "severity": "error",
      "causes": [],
      "url": "https://oxc.rs/docs/guide/usage/linter/rules/typescript/no-explicit-any.html",
      "help": "",
      "filename": "src/types.ts",
      "labels": [
        {
          "span": {
            "offset": 42,
            "length": 3,
            "line": 10,
            "column": 5
          }
        }
      ],
      "related": []
    }
  ],
  "number_of_files": 2,
  "number_of_rules": 93,
  "threads_count": 32,
  "start_time": 0.017410818
}`

func parseTestFindings(t *testing.T, input string) []finding.Finding {
	t.Helper()
	findings, err := new(Detector).parseOutput([]byte(input))
	require.NoError(t, err)
	return findings
}

func TestParseRealOxlintOutput(t *testing.T) {
	t.Parallel()

	findings := parseTestFindings(t, realOxlintOutput)
	require.Len(t, findings, 3)
}

func TestParseFindsCorrectRules(t *testing.T) {
	t.Parallel()

	findings := parseTestFindings(t, realOxlintOutput)

	assert.Equal(t, finding.RuleName("no-debugger"), findings[0].Rule)
	assert.Equal(t, finding.RuleName("no-unused-vars"), findings[1].Rule)
	assert.Equal(t, finding.RuleName("no-explicit-any"), findings[2].Rule)
}

func TestParseFindsCorrectMessages(t *testing.T) {
	t.Parallel()

	findings := parseTestFindings(t, realOxlintOutput)

	assert.Contains(t, findings[0].Message, "debugger")
	assert.Contains(t, findings[1].Message, "declared but never used")
	assert.Contains(t, findings[2].Message, "any")
}

func TestParseFindsCorrectPositions(t *testing.T) {
	t.Parallel()

	findings := parseTestFindings(t, realOxlintOutput)

	assert.Equal(t, finding.FilePath("test.ts"), findings[0].Position.File)
	assert.Equal(t, 2, findings[0].Position.Line)
	assert.Equal(t, 1, findings[0].Position.Column)

	assert.Equal(t, finding.FilePath("src/main.ts"), findings[1].Position.File)
	assert.Equal(t, 1, findings[1].Position.Line)
	assert.Equal(t, 7, findings[1].Position.Column)

	assert.Equal(t, finding.FilePath("src/types.ts"), findings[2].Position.File)
	assert.Equal(t, 10, findings[2].Position.Line)
	assert.Equal(t, 5, findings[2].Position.Column)
}

func TestParseMapsSeverity(t *testing.T) {
	t.Parallel()

	findings := parseTestFindings(t, realOxlintOutput)

	assert.Equal(t, finding.SeverityWarning, findings[0].Severity)
	assert.Equal(t, finding.SeverityWarning, findings[1].Severity)
	assert.Equal(t, finding.SeverityError, findings[2].Severity)
}

func TestParseMapsCategories(t *testing.T) {
	t.Parallel()

	findings := parseTestFindings(t, realOxlintOutput)

	assert.Equal(t, finding.CategoryCorrectness, findings[0].Category)
	assert.Equal(t, finding.CategoryCorrectness, findings[1].Category)
	assert.Equal(t, finding.CategoryTypeSafety, findings[2].Category)
}

func TestParseExtractsURLAndHelp(t *testing.T) {
	t.Parallel()

	findings := parseTestFindings(t, realOxlintOutput)

	assert.Equal(t, "https://oxc.rs/docs/guide/usage/linter/rules/eslint/no-debugger.html",
		findings[0].Metadata["url"])
	assert.Equal(t, "Remove the debugger statement", findings[0].Suggestion)
}

func TestParseEmptyDiagnostics(t *testing.T) {
	t.Parallel()

	findings, err := new(Detector).parseOutput([]byte(`{"diagnostics":[]}`))
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestParseCodeFormat(t *testing.T) {
	t.Parallel()
	tests := []struct {
		code         string
		expectedRule string
		expectedPlug string
	}{
		{"eslint(no-debugger)", "no-debugger", PluginESLint},
		{"typescript/no-explicit-any", "no-explicit-any", PluginTypeScript},
		{"react/exhaustive-deps", "exhaustive-deps", "react"},
		{"unicorn/no-empty-file", "no-empty-file", "unicorn"},
		{"jsx_a11y/alt-text", "alt-text", "jsx_a11y"},
		{"oxc/bad-array-method-on-arguments", "bad-array-method-on-arguments", "oxc"},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			t.Parallel()
			rule, plugin := parseCode(tt.code)
			assert.Equal(t, tt.expectedRule, rule)
			assert.Equal(t, tt.expectedPlug, plugin)
		})
	}
}

func TestDetectEmptyOutput(t *testing.T) {
	t.Parallel()

	_, err := new(Detector).parseOutput([]byte{})
	require.Error(t, err)
}

func TestDetectEmptyJSON(t *testing.T) {
	t.Parallel()

	findings, err := new(Detector).parseOutput([]byte(`{"diagnostics":null}`))
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestDetectOnRealProject(t *testing.T) {
	t.Parallel()
	if os.Getenv("OXLINT_E2E") == "" {
		t.Skip("Set OXLINT_E2E=1 to run e2e test with real oxlint")
	}

	dir := t.TempDir()
	err := os.WriteFile(
		filepath.Join(dir, "test.ts"),
		[]byte("debugger;\nconst x: any = 1;\n"),
		0o644,
	)
	require.NoError(t, err)

	d := NewDetector(dir)
	findings, err := d.Detect(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, findings, "expected findings from test file with debugger and unused var")
}

func TestMapCategory(t *testing.T) {
	t.Parallel()
	tests := []struct {
		plugin   string
		expected finding.Category
	}{
		{PluginESLint, finding.CategoryCorrectness},
		{PluginTypeScript, finding.CategoryTypeSafety},
		{"react", finding.CategoryCorrectness},
		{"react_perf", finding.CategoryCorrectness},
		{"jsx_a11y", finding.CategorySecurity},
		{"unicorn", finding.CategoryStyle},
		{"nextjs", finding.CategoryPerformance},
		{"import", finding.CategoryStructure},
		{"jest", finding.CategoryTesting},
		{"vitest", finding.CategoryTesting},
		{"oxc", finding.CategoryCorrectness},
		{"unknown", finding.CategoryCorrectness},
	}

	for _, tt := range tests {
		t.Run(tt.plugin, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, mapCategory(tt.plugin))
		})
	}
}

func TestPositionFromLabelsEmpty(t *testing.T) {
	t.Parallel()
	pos := positionFromLabels("test.ts", nil)
	assert.Equal(t, finding.FilePath("test.ts"), pos.File)
	assert.Equal(t, 0, pos.Line)
	assert.Equal(t, 0, pos.Column)
	assert.Equal(t, -1, pos.Offset) // unset sentinel per go-finding v1.0.0
}

func TestRangeFromLabels(t *testing.T) {
	t.Parallel()
	t.Run("nil_labels", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, rangeFromLabels("test.ts", nil))
	})

	t.Run("empty_labels", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, rangeFromLabels("test.ts", []oxlintLabel{}))
	})

	t.Run("zero_length", func(t *testing.T) {
		t.Parallel()
		labels := []oxlintLabel{{Span: oxlintSpan{Line: 5, Column: 3, Offset: 42, Length: 0}}}
		assert.Nil(t, rangeFromLabels("test.ts", labels))
	})

	t.Run("single_line_span", func(t *testing.T) {
		t.Parallel()
		labels := []oxlintLabel{{Span: oxlintSpan{Line: 2, Column: 1, Offset: 18, Length: 9}}}
		r := rangeFromLabels("test.ts", labels)
		require.NotNil(t, r)
		assert.Equal(t, 2, r.Start.Line)
		assert.Equal(t, 1, r.Start.Column)
		assert.Equal(t, 18, r.Start.Offset)
		assert.Equal(t, 2, r.End.Line)
		assert.Equal(t, 10, r.End.Column) // 1 + 9
		assert.Equal(t, 27, r.End.Offset) // 18 + 9
		assert.Equal(t, finding.FilePath("test.ts"), r.Start.File)
	})
}

func TestParseOutputPopulatesRange(t *testing.T) {
	t.Parallel()

	findings := parseTestFindings(t, realOxlintOutput)

	require.Len(t, findings, 3)

	r0 := findings[0].Range
	require.NotNil(t, r0, "first finding should have a range")
	assert.Equal(t, 2, r0.Start.Line)
	assert.Equal(t, 1, r0.Start.Column)
	assert.Equal(t, 10, r0.End.Column) // 1 + 9 (length of "debugger")

	r1 := findings[1].Range
	require.NotNil(t, r1, "second finding should have a range")
	assert.Equal(t, 1, r1.Start.Line)
	assert.Equal(t, 7, r1.Start.Column)
	assert.Equal(t, 8, r1.End.Column) // 7 + 1 (length of "x")

	r2 := findings[2].Range
	require.NotNil(t, r2, "third finding should have a range")
	assert.Equal(t, 10, r2.Start.Line)
	assert.Equal(t, 5, r2.Start.Column)
	assert.Equal(t, 8, r2.End.Column) // 5 + 3 (length of "any")
}

func TestJSONRoundTrip(t *testing.T) {
	t.Parallel()

	findings := parseTestFindings(t, realOxlintOutput)

	for _, f := range findings {
		assert.True(t, f.IsValid(), "finding %s should be valid", f.Rule)
	}

	data, err := json.Marshal(findings)
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

// mockRunner returns canned oxlint output for testing.
type mockRunner struct {
	output []byte
	err    error
}

func (m mockRunner) Run(_ context.Context, _ string, _ []string, _ string) ([]byte, error) {
	return m.output, m.err
}

func TestDetectWithMockRunner(t *testing.T) {
	t.Parallel()
	d := NewDetector(".", WithRunner(mockRunner{output: []byte(realOxlintOutput)}))
	findings, err := d.Detect(context.Background())
	require.NoError(t, err)
	assert.Len(t, findings, 3)
	assert.Equal(t, finding.RuleName("no-debugger"), findings[0].Rule)
}

func TestDetectWithMockRunnerEmptyOutput(t *testing.T) {
	t.Parallel()
	d := NewDetector(".", WithRunner(mockRunner{output: []byte("{}")}))
	findings, err := d.Detect(context.Background())
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestDetectorName(t *testing.T) {
	t.Parallel()
	d := NewDetector(".")
	assert.Equal(t, "oxlint", d.Name())
}

func TestWithConfig(t *testing.T) {
	t.Parallel()
	d := NewDetector(".", WithConfig("/path/to/config"))
	assert.Equal(t, "/path/to/config", d.config)
}

func TestWithArgs(t *testing.T) {
	t.Parallel()
	d := NewDetector(".", WithArgs("--verbose"))
	assert.Contains(t, d.args, "--verbose")
}

func TestMapSeverityAllBranches(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected finding.Severity
	}{
		{"error", finding.SeverityError},
		{"warning", finding.SeverityWarning},
		{"warn", finding.SeverityWarning},
		{"info", finding.SeverityInfo},
		{"advice", finding.SeverityInfo},
		{"unknown", finding.SeverityWarning},
		{"", finding.SeverityWarning},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, mapSeverity(tt.input))
		})
	}
}

func TestDetectNonExitError(t *testing.T) {
	t.Parallel()
	d := NewDetector(".", WithRunner(mockRunner{err: assert.AnError}))
	findings, err := d.Detect(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "run oxlint")
	assert.Nil(t, findings)
}

func TestDetectExitErrorWithStderr(t *testing.T) {
	t.Parallel()
	exitErr := &exec.ExitError{Stderr: []byte("something broke")}
	d := NewDetector(".", WithRunner(mockRunner{err: exitErr}))
	findings, err := d.Detect(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "something broke")
	assert.Nil(t, findings)
}

func TestBuildArgsNoConfig(t *testing.T) {
	t.Parallel()
	d := NewDetector("/project")
	args := d.buildArgs()
	assert.Equal(t, []string{"-f", FormatJSON, "."}, args)
}

func TestBuildArgsWithConfig(t *testing.T) {
	t.Parallel()
	d := NewDetector("/project", WithConfig("/project/.oxlintrc.json"))
	args := d.buildArgs()
	assert.Equal(t, []string{"-f", FormatJSON, "-c", "/project/.oxlintrc.json", "."}, args)
}

func TestBuildArgsWithExtraArgs(t *testing.T) {
	t.Parallel()
	d := NewDetector("/project", WithArgs("--verbose", "--silent"))
	args := d.buildArgs()
	assert.Equal(t, []string{"-f", FormatJSON, "--verbose", "--silent", "."}, args)
}

func TestBuildArgsWithConfigAndExtraArgs(t *testing.T) {
	t.Parallel()
	d := NewDetector("/project", WithConfig("/project/.oxlintrc.json"), WithArgs("--verbose"))
	args := d.buildArgs()
	assert.Equal(
		t,
		[]string{"-f", FormatJSON, "-c", "/project/.oxlintrc.json", "--verbose", "."},
		args,
	)
}

func TestParseOutputSetsTag(t *testing.T) {
	t.Parallel()

	findings := parseTestFindings(t, realOxlintOutput)

	assert.Equal(t, PluginESLint, string(findings[0].Tags[0]))
	assert.Equal(t, PluginESLint, string(findings[1].Tags[0]))
	assert.Equal(t, PluginTypeScript, string(findings[2].Tags[0]))
}

func TestParseOutputSetsSnippet(t *testing.T) {
	t.Parallel()

	findings := parseTestFindings(t, realOxlintOutput)

	assert.Empty(t, findings[0].Snippet, "no label text")
	assert.Equal(t, "'x' is declared here", findings[1].Snippet)
}

func TestParseOutputFixStrategyWithoutRegistry(t *testing.T) {
	t.Parallel()

	findings := parseTestFindings(t, realOxlintOutput)

	for _, f := range findings {
		assert.Equal(t, finding.FixStrategyNone, f.FixStrategy,
			"without registry, all findings should have FixStrategyNone")
	}
}

func TestParseOutputFixStrategyWithRegistry(t *testing.T) {
	t.Parallel()

	reg, err := rule.LoadRegistry()
	require.NoError(t, err)

	d := NewDetector(".", WithRegistry(reg))
	findings, err := d.parseOutput([]byte(realOxlintOutput))
	require.NoError(t, err)
	require.Len(t, findings, 3)

	assert.Equal(t, finding.FixStrategyDirect, findings[0].FixStrategy,
		"no-debugger is safe-fixable")
	assert.Equal(t, finding.FixStrategySuggest, findings[2].FixStrategy,
		"no-explicit-any is suggestion-fixable")
}

func TestMapFixStrategy(t *testing.T) {
	t.Parallel()

	reg, err := rule.LoadRegistry()
	require.NoError(t, err)

	d := NewDetector(".", WithRegistry(reg))

	tests := []struct {
		name     string
		rule     string
		plugin   string
		expected finding.FixStrategy
	}{
		{"no-debugger is safe", "no-debugger", PluginESLint, finding.FixStrategyDirect},
		{
			"no-explicit-any is suggestion",
			"no-explicit-any",
			PluginTypeScript,
			finding.FixStrategySuggest,
		},
		{"unknown rule defaults to none", "nonexistent", PluginESLint, finding.FixStrategyNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, d.mapFixStrategy(tt.rule, tt.plugin))
		})
	}
}

func TestMapFixStrategyNoRegistry(t *testing.T) {
	t.Parallel()

	d := NewDetector(".")
	assert.Equal(t, finding.FixStrategyNone, d.mapFixStrategy("no-debugger", PluginESLint))
}

func TestWithRegistry(t *testing.T) {
	t.Parallel()

	reg, err := rule.LoadRegistry()
	require.NoError(t, err)

	d := NewDetector(".", WithRegistry(reg))
	assert.NotNil(t, d.registry)
}
