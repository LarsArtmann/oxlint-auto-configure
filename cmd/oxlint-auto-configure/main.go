// Command oxlint-auto-configure generates optimal oxlint configurations.
package main

import (
	"os"

	"github.com/larsartmann/oxlint-auto-configure/internal/cli"
)

func main() {
	cmd := cli.NewRootCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
