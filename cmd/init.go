package cmd

import (
	"os"

	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize nostromo configuration",
	Long: `Create a nostromo config file with defaults.

By default the config file is located at ~/.nostromo/ships/manifest.yaml.

Customize this with the $NOSTROMO_HOME environment variable`,
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(task.InitConfig(cmd.Root()))
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
