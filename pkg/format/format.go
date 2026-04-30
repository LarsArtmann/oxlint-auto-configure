// Package format renders analysis findings in various output formats.
package format

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// FindingView is a projection of a lint finding for rendering.
type FindingView struct {
	Rule        string `json:"rule"`
	Message     string `json:"message"`
	Severity    string `json:"severity"`
	Category    string `json:"category"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	Column      int    `json:"column"`
	DocsURL     string `json:"docs_url,omitempty"`
	FixStrategy string `json:"fix_strategy,omitempty"`
	Tag         string `json:"tag,omitempty"`
	Snippet     string `json:"snippet,omitempty"`
}

// SummaryView is a projection of analysis summary statistics.
type SummaryView struct {
	Total         int
	BySeverity    map[string]int
	ByCategory    map[string]int
	FilesAffected int
	Iterations    int
	Stable        bool
}

// PrintSummary writes a human-readable summary to w.
func PrintSummary(w io.Writer, sv *SummaryView) error {
	_, _ = fmt.Fprintf(
		w,
		"findings: total=%d, by_severity=%s, by_category=%s, files_affected=%d, iterations=%d, stable=%t\n",
		sv.Total,
		formatMap(sv.BySeverity),
		formatMap(sv.ByCategory),
		sv.FilesAffected,
		sv.Iterations,
		sv.Stable,
	)
	return nil
}

// PrintFindingsJSON writes findings as a JSON array to w.
func PrintFindingsJSON(w io.Writer, findings []FindingView) error {
	data, err := json.MarshalIndent(findings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal findings: %w", err)
	}
	_, _ = fmt.Fprintln(w, string(data))
	return nil
}

// PrintFindingsTable writes findings as a Markdown table to w.
func PrintFindingsTable(w io.Writer, findings []FindingView) error {
	_, _ = fmt.Fprintln(w, "| Rule | Severity | Category | File:Line | Message |")
	_, _ = fmt.Fprintln(w, "|------|----------|----------|-----------|---------|")

	sorted := make([]FindingView, len(findings))
	copy(sorted, findings)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].File != sorted[j].File {
			return sorted[i].File < sorted[j].File
		}
		return sorted[i].Line < sorted[j].Line
	})

	for _, f := range sorted {
		loc := fmt.Sprintf("%s:%d", f.File, f.Line)
		msg := f.Message
		if len(msg) > 60 {
			msg = msg[:57] + "..."
		}
		msg = strings.ReplaceAll(msg, "|", "\\|")
		_, _ = fmt.Fprintf(w, "| %s | %s | %s | %s | %s |\n",
			f.Rule, f.Severity, f.Category, loc, msg)
	}
	return nil
}

// Map returns a sorted "key=count, ..." string for a map.
func Map(m map[string]int) string {
	return formatMap(m)
}

func formatMap(m map[string]int) string {
	parts := make([]string, 0, len(m))
	for k, v := range m {
		parts = append(parts, fmt.Sprintf("%s=%d", k, v))
	}
	sort.Strings(parts)
	return strings.Join(parts, ", ")
}
