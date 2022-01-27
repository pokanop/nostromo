package cmd

import (
	"os"

	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

var destroyAll bool

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy nostromo configuration",
	Long: `Destroy nostromo configuration.

By default the core manifest is located at ~/.nostromo/ships/manifest.yaml.
Optionally delete all config files using --all flag.

Customize this with the $NOSTROMO_HOME environment variable`,
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(task.DestroyConfig(destroyAll))
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)

	destroyCmd.Flags().BoolVarP(&destroyAll, "all", "a", false, "Delete all configuration files")
}
