// Package oxlint provides a go-finding Detector that runs oxlint and
// converts its JSON output to go-finding Finding values.
package oxlint

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	finding "github.com/larsartmann/go-finding"
	"github.com/larsartmann/oxlint-auto-configure/pkg/rule"
)

// Runner executes an oxlint command and returns its stdout.
type Runner interface {
	Run(ctx context.Context, name string, args []string, dir string) ([]byte, error)
}

// Plugin name constants.
const (
	PluginESLint     = "eslint"
	PluginTypeScript = "typescript"
	FormatJSON       = "json"
)

// realRunner executes oxlint via exec.CommandContext.
type realRunner struct{}

func (realRunner) Run(ctx context.Context, name string, args []string, dir string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	return out, finding.NewIOError("run oxlint", err)
}

// Detector runs oxlint and converts findings to the go-finding model.
type Detector struct {
	rootDir  string
	config   string
	args     []string
	runner   Runner
	registry *rule.Registry
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

// WithRunner sets the command runner (for testing).
func WithRunner(r Runner) Option {
	return func(d *Detector) { d.runner = r }
}

// WithRegistry sets the rule registry for fix-strategy lookup.
func WithRegistry(reg *rule.Registry) Option {
	return func(d *Detector) { d.registry = reg }
}

// NewDetector creates an oxlint detector for the given root directory.
func NewDetector(rootDir string, opts ...Option) *Detector {
	d := &Detector{rootDir: rootDir, runner: realRunner{}}
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

	output, err := d.runner.Run(ctx, "oxlint", args, d.rootDir)
	if err := checkExitError(err, "oxlint"); err != nil {
		return nil, err
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
	Message  string        `json:"message"`
	Code     string        `json:"code"`
	Severity string        `json:"severity"`
	Filename string        `json:"filename"`
	Labels   []oxlintLabel `json:"labels"`
	URL      string        `json:"url"`
	Help     string        `json:"help"`
}

// oxlintLabel represents a labeled span in the source code.
type oxlintLabel struct {
	Label string     `json:"label,omitempty"`
	Span  oxlintSpan `json:"span"`
}

// oxlintSpan represents a source range.
type oxlintSpan struct {
	Offset int `json:"offset"`
	Length int `json:"length"`
	Line   int `json:"line"`
	Column int `json:"column"`
}

// parseOutput converts raw oxlint JSON into go-finding findings.
func (d *Detector) parseOutput(data []byte) ([]finding.Finding, error) {
	var output oxlintOutput
	if err := json.Unmarshal(data, &output); err != nil {
		return nil, finding.NewParseError("oxlint JSON", err)
	}

	findings := make([]finding.Finding, 0, len(output.Diagnostics))
	for _, diag := range output.Diagnostics {
		ruleName, pluginName := parseCode(diag.Code)

		f := finding.NewFinding(
			finding.RuleName(ruleName),
			finding.ToolName("oxlint"),
			diag.Message,
			mapSeverity(diag.Severity),
			positionFromLabels(diag.Filename, diag.Labels),
			1.0,
		)

		f.Category = mapCategory(pluginName)
		f.Range = rangeFromLabels(diag.Filename, diag.Labels)
		f.FixStrategy = d.mapFixStrategy(ruleName, pluginName)
		f.Tags = []finding.Tag{finding.Tag(pluginName)}

		if diag.URL != "" {
			if f.Metadata == nil {
				f.Metadata = make(map[string]string)
			}
			f.Metadata["url"] = diag.URL
		}
		if diag.Help != "" {
			f.Suggestion = diag.Help
		}
		if len(diag.Labels) > 0 && diag.Labels[0].Label != "" {
			f.Snippet = diag.Labels[0].Label
		}

		findings = append(findings, f)
	}

	return findings, nil
}

// positionFromLabels builds a Position from the first label's span.
// When no labels are present, returns a position with Offset=-1 (unset sentinel)
// per go-finding v1.0.0 zero-value semantics.
func positionFromLabels(filename string, labels []oxlintLabel) finding.Position {
	file := finding.FilePath(filename)
	if len(labels) == 0 {
		return finding.Position{File: file, Offset: -1}
	}
	return finding.Position{
		File:   file,
		Line:   labels[0].Span.Line,
		Column: labels[0].Span.Column,
		Offset: labels[0].Span.Offset,
	}
}

// rangeFromLabels constructs a Range from the first label's span.
// Uses offset+length to compute end position. For single-line spans
// (the common case), end column is start column + length.
// Returns nil if no labels are present.
func rangeFromLabels(filename string, labels []oxlintLabel) *finding.Range {
	file := finding.FilePath(filename)
	if len(labels) == 0 {
		return nil
	}

	span := labels[0].Span
	start := finding.Position{
		File:   file,
		Line:   span.Line,
		Column: span.Column,
		Offset: span.Offset,
	}

	if span.Length == 0 {
		return nil
	}

	end := finding.Position{
		File:   file,
		Line:   span.Line,
		Column: span.Column + span.Length,
		Offset: span.Offset + span.Length,
	}

	return &finding.Range{Start: start, End: end}
}

// parseCode splits "eslint(no-debugger)" or "typescript/no-explicit-any"
// into (rule-name, plugin).
func parseCode(code string) (ruleName, plugin string) {
	plugin, ruleName, found := strings.Cut(code, "(")
	if found {
		return strings.TrimSuffix(ruleName, ")"), plugin
	}
	plugin, ruleName, found = strings.Cut(code, "/")
	if found {
		return ruleName, plugin
	}
	return code, PluginESLint
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

// pluginToCategory maps oxlint plugins to go-finding categories.
var pluginToCategory = map[rule.Plugin]finding.Category{ //nolint:gochecknoglobals // immutable lookup table
	rule.PluginTypeScript: finding.CategoryTypeSafety,
	rule.PluginReact:      finding.CategoryCorrectness,
	rule.PluginReactPerf:  finding.CategoryCorrectness,
	rule.PluginJSXA11y:    finding.CategorySecurity,
	rule.PluginUnicorn:    finding.CategoryStyle,
	rule.PluginNextJS:     finding.CategoryPerformance,
	rule.PluginImport:     finding.CategoryStructure,
	rule.PluginJest:       finding.CategoryTesting,
	rule.PluginVitest:     finding.CategoryTesting,
	rule.PluginOXC:        finding.CategoryCorrectness,
	rule.PluginESLint:     finding.CategoryCorrectness,
}

func mapCategory(pluginName string) finding.Category {
	if cat, ok := pluginToCategory[rule.Plugin(pluginName)]; ok {
		return cat
	}
	return finding.CategoryCorrectness
}

// mapFixStrategy returns the go-finding FixStrategy for the given rule.
// Uses the registry when available; falls back to FixStrategyNone.
func (d *Detector) mapFixStrategy(ruleName, pluginName string) finding.FixStrategy {
	if d.registry == nil {
		return finding.FixStrategyNone
	}

	fullName := ruleName
	if pluginName != PluginESLint {
		fullName = pluginName + "/" + ruleName
	}

	r, ok := d.registry.ByName(fullName)
	if !ok {
		return finding.FixStrategyNone
	}

	switch r.Fix {
	case rule.FixSafe:
		return finding.FixStrategyDirect
	case rule.FixSuggestion, rule.FixDangerous:
		return finding.FixStrategySuggest
	case rule.FixNone:
		return finding.FixStrategyNone
	default:
		return finding.FixStrategyNone
	}
}

// handleExitError returns a meaningful error for exec.ExitError, or nil
// if the exit code is just oxlint reporting findings (exit code 1).
// Returns a non-nil error for unexpected failures.
func handleExitError(err error, cmd string) error {
	exitErr := &exec.ExitError{}
	if errors.As(err, &exitErr) {
		if len(exitErr.Stderr) > 0 {
			return finding.NewIOError(cmd, fmt.Errorf("%s", string(exitErr.Stderr)))
		}
		return nil
	}
	return finding.NewIOError("run "+cmd, err)
}

// checkExitError checks the error and returns it if it's a real error (not just findings).
func checkExitError(err error, cmd string) error {
	if err != nil {
		if exitErr := handleExitError(err, cmd); exitErr != nil {
			return exitErr
		}
	}
	return nil
}
