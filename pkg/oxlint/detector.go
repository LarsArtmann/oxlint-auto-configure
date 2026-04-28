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
func (d *Detector) Detect(ctx context.Context) ([]finding.Finding, error) {
	args := d.buildArgs()
	cmd := exec.CommandContext(ctx, "oxlint", args...)
	cmd.Dir = d.rootDir

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if len(exitErr.Stderr) > 0 {
				return nil, fmt.Errorf("oxlint stderr: %s", string(exitErr.Stderr))
			}
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

// oxlintJSON represents the JSON output from oxlint.
type oxlintJSON struct {
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Rule     string `json:"rule"`
	Filename string `json:"filename"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Fix      bool   `json:"fix"`
}

func (d *Detector) parseOutput(data []byte) ([]finding.Finding, error) {
	var results []oxlintJSON
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, fmt.Errorf("parse oxlint JSON: %w", err)
	}

	findings := make([]finding.Finding, 0, len(results))
	for _, r := range results {
		f := finding.NewFinding(
			r.Rule,
			"oxlint",
			r.Message,
			mapSeverity(r.Severity),
			finding.Position{File: r.Filename, Line: r.Line, Column: r.Column},
		)

		if r.Fix {
			f.FixStrategy = finding.FixStrategyDirect
		}

		f.Category = mapCategory(r.Rule)
		findings = append(findings, f)
	}

	return findings, nil
}

func mapSeverity(s string) finding.Severity {
	switch strings.ToLower(s) {
	case "error":
		return finding.SeverityError
	case "warning", "warn":
		return finding.SeverityWarning
	case "info":
		return finding.SeverityInfo
	default:
		return finding.SeverityWarning
	}
}

func mapCategory(rule string) finding.Category {
	parts := strings.SplitN(rule, "/", 2)
	prefix := ""
	if len(parts) == 2 {
		prefix = parts[0]
	}

	switch prefix {
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
	default:
		return finding.CategoryCorrectness
	}
}
