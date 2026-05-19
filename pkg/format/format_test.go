package format

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/larsartmann/oxlint-auto-configure/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testRuleNoDebugger      = "no-debugger"
	testCategoryCorrectness = "correctness"
)

func TestPrintSummary(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	sv := &SummaryView{
		Total:         5,
		BySeverity:    map[string]int{config.SeverityError: 3, "warning": 2},
		ByCategory:    map[string]int{testCategoryCorrectness: 5},
		ByFixStrategy: map[string]int{"none": 3, "suggest": 2},
		FilesAffected: 3,
		Iterations:    1,
		Stable:        true,
		Findings: []FindingView{
			{Rule: "no-unused-vars", File: "a.ts", Line: 1},
			{Rule: "no-unused-vars", File: "b.ts", Line: 2},
			{Rule: "no-console", File: "a.ts", Line: 3},
			{Rule: testRuleNoDebugger, File: "c.ts", Line: 4},
			{Rule: "eqeqeq", File: "d.ts", Line: 5},
		},
	}

	err := PrintSummary(&buf, sv)
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "5 finding(s) across 3 file(s)")
	assert.Contains(t, output, "error")
	assert.Contains(t, output, "3")
	assert.Contains(t, output, "warning")
	assert.Contains(t, output, "2")
	assert.Contains(t, output, "correctness")
	assert.Contains(t, output, "Top rules")
	assert.Contains(t, output, "Top files")
	assert.Contains(t, output, "no-unused-vars")
}

func TestPrintFindingsJSON(t *testing.T) {
	t.Parallel()
	findings := []FindingView{
		{
			Rule:     testRuleNoDebugger,
			Severity: config.SeverityError,
			Category: testCategoryCorrectness,
			File:     "test.js",
			Line:     10,
			Column:   5,
			Message:  "Unexpected debugger statement",
		},
	}

	var buf bytes.Buffer
	err := PrintFindingsJSON(&buf, findings)
	require.NoError(t, err)

	var parsed []FindingView
	err = json.Unmarshal(buf.Bytes(), &parsed)
	require.NoError(t, err)
	assert.Len(t, parsed, 1)
	assert.Equal(t, testRuleNoDebugger, parsed[0].Rule)
}

func TestPrintFindingsTable(t *testing.T) {
	t.Parallel()
	findings := []FindingView{
		{
			Rule:     testRuleNoDebugger,
			Severity: config.SeverityError,
			Category: testCategoryCorrectness,
			File:     "a.js",
			Line:     5,
			Column:   1,
			Message:  "debugger",
		},
		{
			Rule:     "no-console",
			Severity: "warn",
			Category: "style",
			File:     "b.js",
			Line:     3,
			Column:   1,
			Message:  "console.log",
		},
	}

	var buf bytes.Buffer
	err := PrintFindingsTable(&buf, findings)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "| Rule |")
	assert.Contains(t, output, "no-debugger")
	assert.Contains(t, output, "no-console")
}

func TestPrintFindingsTableSortsByFile(t *testing.T) {
	t.Parallel()
	findings := []FindingView{
		{Rule: "a-rule", File: "a.js", Line: 1},
		{Rule: "z-rule", File: "z.js", Line: 1},
	}

	var buf bytes.Buffer
	err := PrintFindingsTable(&buf, findings)
	require.NoError(t, err)

	lines := strings.Split(buf.String(), "\n")
	aIdx := 0
	zIdx := 0
	for i, line := range lines {
		if strings.Contains(line, "a-rule") {
			aIdx = i
		}
		if strings.Contains(line, "z-rule") {
			zIdx = i
		}
	}
	assert.Less(t, aIdx, zIdx, "a.js should appear before z.js (caller must pre-sort)")
}

func TestPrintFindingsTableTruncatesLongMessages(t *testing.T) {
	t.Parallel()
	longMsg := strings.Repeat("x", 100)
	findings := []FindingView{
		{
			Rule:     "test",
			Severity: config.SeverityError,
			Category: testCategoryCorrectness,
			File:     "f.js",
			Line:     1,
			Message:  longMsg,
		},
	}

	var buf bytes.Buffer
	err := PrintFindingsTable(&buf, findings)
	require.NoError(t, err)

	assert.Contains(t, buf.String(), "...")
}

func TestFormatMap(t *testing.T) {
	t.Parallel()
	m := map[string]int{"warn": 5, "error": 3}
	result := Map(m)
	assert.Contains(t, result, "error=3")
	assert.Contains(t, result, "warn=5")
}
