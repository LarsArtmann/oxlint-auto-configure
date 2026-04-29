package format

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrintSummary(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	sv := &SummaryView{
		Total:         5,
		BySeverity:    map[string]int{"error": 3, "warn": 2},
		ByCategory:    map[string]int{"correctness": 5},
		FilesAffected: 3,
		Iterations:    1,
		Stable:        true,
	}

	err := PrintSummary(&buf, sv)
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "total=5")
	assert.Contains(t, output, "error=3")
	assert.Contains(t, output, "files_affected=3")
	assert.Contains(t, output, "stable=true")
}

func TestPrintFindingsJSON(t *testing.T) {
	t.Parallel()
	findings := []FindingView{
		{
			Rule:     "no-debugger",
			Severity: "error",
			Category: "correctness",
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
	assert.Equal(t, "no-debugger", parsed[0].Rule)
}

func TestPrintFindingsTable(t *testing.T) {
	t.Parallel()
	findings := []FindingView{
		{
			Rule:     "no-debugger",
			Severity: "error",
			Category: "correctness",
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
		{Rule: "z-rule", File: "z.js", Line: 1},
		{Rule: "a-rule", File: "a.js", Line: 1},
	}

	var buf bytes.Buffer
	err := PrintFindingsTable(&buf, findings)
	require.NoError(t, err)

	lines := strings.Split(buf.String(), "\n")
	aLine := 0
	zLine := 0
	for i, line := range lines {
		if strings.Contains(line, "a-rule") {
			aLine = i
		}
		if strings.Contains(line, "z-rule") {
			zLine = i
		}
	}
	assert.Less(t, aLine, zLine, "a.js should appear before z.js")
}

func TestPrintFindingsTableTruncatesLongMessages(t *testing.T) {
	t.Parallel()
	longMsg := strings.Repeat("x", 100)
	findings := []FindingView{
		{
			Rule:     "test",
			Severity: "error",
			Category: "correctness",
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
