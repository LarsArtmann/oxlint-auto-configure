package oxlint

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	finding "github.com/larsartmann/go-finding"
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

func TestParseRealOxlintOutput(t *testing.T) {
	t.Parallel()

	findings, err := parseOutput([]byte(realOxlintOutput))
	require.NoError(t, err)
	require.Len(t, findings, 3)
}

func TestParseFindsCorrectRules(t *testing.T) {
	t.Parallel()

	findings, err := parseOutput([]byte(realOxlintOutput))
	require.NoError(t, err)

	assert.Equal(t, "no-debugger", findings[0].Rule)
	assert.Equal(t, "no-unused-vars", findings[1].Rule)
	assert.Equal(t, "no-explicit-any", findings[2].Rule)
}

func TestParseFindsCorrectMessages(t *testing.T) {
	t.Parallel()

	findings, err := parseOutput([]byte(realOxlintOutput))
	require.NoError(t, err)

	assert.Contains(t, findings[0].Message, "debugger")
	assert.Contains(t, findings[1].Message, "declared but never used")
	assert.Contains(t, findings[2].Message, "any")
}

func TestParseFindsCorrectPositions(t *testing.T) {
	t.Parallel()

	findings, err := parseOutput([]byte(realOxlintOutput))
	require.NoError(t, err)

	assert.Equal(t, "test.ts", findings[0].Position.File)
	assert.Equal(t, 2, findings[0].Position.Line)
	assert.Equal(t, 1, findings[0].Position.Column)

	assert.Equal(t, "src/main.ts", findings[1].Position.File)
	assert.Equal(t, 1, findings[1].Position.Line)
	assert.Equal(t, 7, findings[1].Position.Column)

	assert.Equal(t, "src/types.ts", findings[2].Position.File)
	assert.Equal(t, 10, findings[2].Position.Line)
	assert.Equal(t, 5, findings[2].Position.Column)
}

func TestParseMapsSeverity(t *testing.T) {
	t.Parallel()

	findings, err := parseOutput([]byte(realOxlintOutput))
	require.NoError(t, err)

	assert.Equal(t, finding.SeverityWarning, findings[0].Severity)
	assert.Equal(t, finding.SeverityWarning, findings[1].Severity)
	assert.Equal(t, finding.SeverityError, findings[2].Severity)
}

func TestParseMapsCategories(t *testing.T) {
	t.Parallel()

	findings, err := parseOutput([]byte(realOxlintOutput))
	require.NoError(t, err)

	assert.Equal(t, finding.CategoryCorrectness, findings[0].Category)
	assert.Equal(t, finding.CategoryCorrectness, findings[1].Category)
	assert.Equal(t, finding.CategoryTypeSafety, findings[2].Category)
}

func TestParseExtractsURLAndHelp(t *testing.T) {
	t.Parallel()

	findings, err := parseOutput([]byte(realOxlintOutput))
	require.NoError(t, err)

	assert.Equal(t, "https://oxc.rs/docs/guide/usage/linter/rules/eslint/no-debugger.html",
		findings[0].Metadata["url"])
	assert.Equal(t, "Remove the debugger statement", findings[0].Suggestion)
}

func TestParseEmptyDiagnostics(t *testing.T) {
	t.Parallel()

	findings, err := parseOutput([]byte(`{"diagnostics":[]}`))
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
		{"eslint(no-debugger)", "no-debugger", "eslint"},
		{"typescript/no-explicit-any", "no-explicit-any", "typescript"},
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

	_, err := parseOutput([]byte{})
	require.Error(t, err)
}

func TestDetectEmptyJSON(t *testing.T) {
	t.Parallel()

	findings, err := parseOutput([]byte(`{"diagnostics":null}`))
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
		{"eslint", finding.CategoryCorrectness},
		{"typescript", finding.CategoryTypeSafety},
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
	line, col := positionFromLabels(nil)
	assert.Equal(t, 0, line)
	assert.Equal(t, 0, col)
}

func TestJSONRoundTrip(t *testing.T) {
	t.Parallel()

	findings, err := parseOutput([]byte(realOxlintOutput))
	require.NoError(t, err)

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
	assert.Equal(t, "no-debugger", findings[0].Rule)
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
