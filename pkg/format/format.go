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
	ByFixStrategy map[string]int
	FilesAffected int
	Iterations    int
	Stable        bool
	Findings      []FindingView
}

// PrintSummary writes a human-readable summary to w.
func PrintSummary(w io.Writer, sv *SummaryView) error {
	fmt.Fprintf(w, "\n=== Analysis Results ===\n")
	fmt.Fprintf(w, "\n%d finding(s) across %d file(s)", sv.Total, sv.FilesAffected)
	if sv.Iterations > 1 {
		fmt.Fprintf(w, " (%d iterations, stable=%t)", sv.Iterations, sv.Stable)
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, "\nBy severity:")
	for _, sev := range sortedKeys(sv.BySeverity) {
		fmt.Fprintf(w, "  %-10s %d\n", sev, sv.BySeverity[sev])
	}

	fmt.Fprintln(w, "\nBy category:")
	for _, cat := range sortedKeys(sv.ByCategory) {
		fmt.Fprintf(w, "  %-15s %d\n", cat, sv.ByCategory[cat])
	}

	if len(sv.ByFixStrategy) > 0 {
		fmt.Fprintln(w, "\nBy fix strategy:")
		for _, fix := range sortedKeys(sv.ByFixStrategy) {
			fmt.Fprintf(w, "  %-15s %d\n", fix, sv.ByFixStrategy[fix])
		}
	}

	topRules := topByRule(sv.Findings, 10)
	if len(topRules) > 0 {
		fmt.Fprintln(w, "\nTop rules:")
		for _, entry := range topRules {
			fmt.Fprintf(w, "  %-45s %d\n", entry.name, entry.count)
		}
	}

	topFiles := topByFile(sv.Findings, 10)
	if len(topFiles) > 0 {
		fmt.Fprintln(w, "\nTop files:")
		for _, entry := range topFiles {
			fmt.Fprintf(w, "  %-45s %d\n", entry.name, entry.count)
		}
	}

	fmt.Fprintln(w)
	return nil
}

type namedCount struct {
	name  string
	count int
}

func topByRule(findings []FindingView, n int) []namedCount {
	counts := make(map[string]int)
	for _, f := range findings {
		counts[f.Rule]++
	}
	return topN(counts, n)
}

func topByFile(findings []FindingView, n int) []namedCount {
	counts := make(map[string]int)
	for _, f := range findings {
		counts[f.File]++
	}
	return topN(counts, n)
}

func topN(counts map[string]int, n int) []namedCount {
	entries := make([]namedCount, 0, len(counts))
	for k, v := range counts {
		entries = append(entries, namedCount{k, v})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].count != entries[j].count {
			return entries[i].count > entries[j].count
		}
		return entries[i].name < entries[j].name
	})
	if len(entries) > n {
		entries = entries[:n]
	}
	return entries
}

func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
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
// Findings should be pre-sorted by the caller (e.g., finding.SortByPosition).
func PrintFindingsTable(w io.Writer, findings []FindingView) error {
	_, _ = fmt.Fprintln(w, "| Rule | Severity | Category | File:Line | Message |")
	_, _ = fmt.Fprintln(w, "|------|----------|----------|-----------|---------|")

	for _, f := range findings {
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
