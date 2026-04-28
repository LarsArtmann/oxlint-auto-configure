// Package oxlint provides a go-finding Detector that runs oxlint and
// converts its JSON output to go-finding Finding values.
package oxlint

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	finding "github.com/larsartmann/go-finding"
)

// Detector runs oxlint and converts findings to the go-finding model.
type Detector struct {
	rootDir string
	config  string
	args    []string
}

// Option configures the oxlint detector.
type Option func(*Detector)

// WithConfig sets the oxlint config file path.
func WithConfig(path string) Option {
	return func(d *Detector) { d.config = path }
}

// WithArgs adds extra arguments to the oxlint invocation.
func WithArgs(args ...string) Option {
	return func(d *Detector) { d.args = append(d.args, args...) }
}

// NewDetector creates an oxlint detector for the given root directory.
func NewDetector(rootDir string, opts ...Option) *Detector {
	d := &Detector{rootDir: rootDir}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

// Name implements finding.Detector.
func (d *Detector) Name() string { return "oxlint" }

// Detect runs oxlint and returns findings.
// oxlint exits with code 1 when it finds issues, which is not an error.
func (d *Detector) Detect(ctx context.Context) ([]finding.Finding, error) {
	args := d.buildArgs()
	cmd := exec.CommandContext(ctx, "oxlint", args...)
	cmd.Dir = d.rootDir

	output, err := cmd.Output()
	if err != nil {
		// oxlint exits non-zero when findings exist — that's not an error for us.
		// Only surface real failures (stderr output, command not found, etc.)
		if exitErr, ok := err.(*exec.ExitError); ok {
			// If there's stderr, that's a real error
			if len(exitErr.Stderr) > 0 {
				return nil, fmt.Errorf("oxlint: %s", string(exitErr.Stderr))
			}
			// Exit code 1 with no stderr = findings found, not a failure
		} else {
			// Non-ExitError (e.g., oxlint not found in PATH)
			return nil, fmt.Errorf("run oxlint: %w", err)
		}
	}

	if len(output) == 0 {
		return nil, nil
	}

	return d.parseOutput(output)
}

func (d *Detector) buildArgs() []string {
	var args []string
	args = append(args, "-f", "json")

	if d.config != "" {
		args = append(args, "-c", d.config)
	}

	args = append(args, d.args...)
	args = append(args, ".")
	return args
}

// oxlintOutput represents the top-level oxlint JSON output.
type oxlintOutput struct {
	Diagnostics []oxlintDiagnostic `json:"diagnostics"`
}

// oxlintDiagnostic represents a single diagnostic from oxlint.
type oxlintDiagnostic struct {
	Message  string          `json:"message"`
	Code     string          `json:"code"`
	Severity string          `json:"severity"`
	Filename string          `json:"filename"`
	Labels   []oxlintLabel   `json:"labels"`
	URL      string          `json:"url"`
	Help     string          `json:"help"`
}

// oxlintLabel represents a labeled span in the source code.
type oxlintLabel struct {
	Label string    `json:"label,omitempty"`
	Span  oxlintSpan `json:"span"`
}

// oxlintSpan represents a source range.
type oxlintSpan struct {
	Offset int `json:"offset"`
	Length int `json:"length"`
	Line   int `json:"line"`
	Column int `json:"column"`
}

func (d *Detector) parseOutput(data []byte) ([]finding.Finding, error) {
	var output oxlintOutput
	if err := json.Unmarshal(data, &output); err != nil {
		return nil, fmt.Errorf("parse oxlint JSON: %w", err)
	}

	findings := make([]finding.Finding, 0, len(output.Diagnostics))
	for _, diag := range output.Diagnostics {
		line, col := positionFromLabels(diag.Labels)
		ruleName, pluginName := parseCode(diag.Code)

		f := finding.NewFinding(
			ruleName,
			"oxlint",
			diag.Message,
			mapSeverity(diag.Severity),
			finding.Position{File: diag.Filename, Line: line, Column: col},
		)

		f.Category = mapCategory(pluginName)

		if diag.URL != "" {
			f.Metadata = map[string]string{"url": diag.URL}
		}
		if diag.Help != "" {
			f.Suggestion = diag.Help
		}

		findings = append(findings, f)
	}

	return findings, nil
}

// positionFromLabels extracts line/column from the first label's span.
func positionFromLabels(labels []oxlintLabel) (line, col int) {
	if len(labels) > 0 {
		return labels[0].Span.Line, labels[0].Span.Column
	}
	return 0, 0
}

// parseCode splits "eslint(no-debugger)" or "typescript/no-explicit-any"
// into (rule-name, plugin).
func parseCode(code string) (ruleName, plugin string) {
	// Format 1: "plugin(rule-name)" e.g. "eslint(no-debugger)"
	if idx := strings.IndexByte(code, '('); idx >= 0 {
		plugin = code[:idx]
		ruleName = strings.TrimSuffix(code[idx+1:], ")")
		return ruleName, plugin
	}
	// Format 2: "plugin/rule-name" e.g. "typescript/no-explicit-any"
	if idx := strings.IndexByte(code, '/'); idx >= 0 {
		plugin = code[:idx]
		ruleName = code[idx+1:]
		return ruleName, plugin
	}
	return code, "eslint"
}

func mapSeverity(s string) finding.Severity {
	switch strings.ToLower(s) {
	case "error":
		return finding.SeverityError
	case "warning", "warn":
		return finding.SeverityWarning
	case "info", "advice":
		return finding.SeverityInfo
	default:
		return finding.SeverityWarning
	}
}

func mapCategory(plugin string) finding.Category {
	switch plugin {
	case "typescript":
		return finding.CategoryTypeSafety
	case "react", "react_perf":
		return finding.CategoryCorrectness
	case "jsx_a11y":
		return finding.CategorySecurity
	case "unicorn":
		return finding.CategoryStyle
	case "nextjs":
		return finding.CategoryPerformance
	case "import":
		return finding.CategoryStructure
	case "jest", "vitest":
		return finding.CategoryTesting
	case "oxc":
		return finding.CategoryCorrectness
	default:
		return finding.CategoryCorrectness
	}
}
