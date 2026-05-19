package cli

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"sync"

	"github.com/larsartmann/oxlint-auto-configure/pkg/profile"
	"github.com/spf13/cobra"
)

var logSetupMu sync.Mutex //nolint:gochecknoglobals // protects concurrent slog setup

const (
	defaultConfigPath = ".oxlintrc.json"
	defaultProfile    = profile.ProfileRecommended
)

// Command names (used across cmd files and tests).
const (
	CmdConfigure = "configure"
	CmdReport    = "report"
	CmdValidate  = "validate"
	CmdAnalyze   = "analyze"
)

// Output formats for analyze and report commands.
const (
	FormatSummary = "summary"
	FormatReport  = "report"
	FormatJSON    = "json"
	FormatTable   = "table"
	FormatSARIF   = "sarif"
)

var version = "dev"

// NewRootCommand creates the root CLI command.
func NewRootCommand() *cobra.Command {
	var (
		verbose bool
		quiet   bool
	)

	root := &cobra.Command{
		Use:   "oxlint-auto-configure",
		Short: "Automatically configure oxlint for maximum type safety",
		Long: `oxlint-auto-configure generates the optimal .oxlintrc.json for your project.

This tool configures oxlint — it does not replace it. Running oxlint,
auto-fixing code, and enforcing rules are oxlint's job.

It detects your project type, picks the right plugins, and sets every
rule to the best severity based on your chosen profile.`,
		Version: version,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			return setupLogging(verbose, quiet, os.Stderr)
		},
	}

	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	root.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress non-error output")

	root.AddCommand(newConfigureCommand())
	root.AddCommand(newAnalyzeCommand())
	root.AddCommand(newValidateCommand())
	root.AddCommand(newReportCommand())

	return root
}

func setupLogging(verbose, quiet bool, w io.Writer) error {
	logSetupMu.Lock()
	defer logSetupMu.Unlock()

	if verbose && quiet {
		return errors.New("cannot use both --verbose and --quiet")
	}

	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	if quiet {
		level = slog.LevelError
	}

	opts := &slog.HandlerOptions{
		Level:       level,
		ReplaceAttr: compactLogAttr,
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(w, opts)))
	return nil
}

// compactLogAttr strips the timestamp from log output for cleaner CLI.
func compactLogAttr(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey {
		return slog.Attr{}
	}
	return a
}
