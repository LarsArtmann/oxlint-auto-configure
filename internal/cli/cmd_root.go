package cli

import (
	"github.com/larsartmann/oxlint-auto-configure/pkg/profile"
	"github.com/spf13/cobra"
)

const (
	defaultConfigPath = ".oxlintrc.json"
	defaultProfile   = profile.ProfileRecommended
)

var version = "dev"

func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "oxlint-auto-configure",
		Short: "Automatically configure oxlint for maximum type safety",
		Long: `oxlint-auto-configure analyzes your project and generates the optimal
.oxlintrc.json configuration for maximum type safety and correctness enforcement.

It uses the go-finding library to run oxlint, collect findings, and
auto-configure every available rule with the best severity setting.`,
		Version: version,
	}

	root.AddCommand(newConfigureCommand())
	root.AddCommand(newAnalyzeCommand())
	root.AddCommand(newValidateCommand())
	root.AddCommand(newReportCommand())

	return root
}
