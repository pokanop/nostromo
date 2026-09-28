package cmd

import (
	"os"

	"github.com/pokanop/nostromo/banner"
	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

var noBanner bool

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize nostromo configuration",
	Long: `Create a nostromo config file with defaults.

By default the config file is located at ~/.nostromo/ships/manifest.yaml.

Customize this with the $NOSTROMO_HOME environment variable

A welcome banner is shown the first time a config is created when running
in a terminal. Pass --no-banner to suppress it (e.g. in scripts).`,
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(task.InitConfig(cmd.Root(), showBanner(noBanner, os.Stdout)))
	},
}

// showBanner reports whether the init banner should be printed: only when
// not disabled and stdout is an interactive terminal.
func showBanner(disabled bool, stdout *os.File) bool {
	return !disabled && banner.IsTerminal(stdout)
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().BoolVar(&noBanner, "no-banner", false, "do not print the welcome banner")
}
