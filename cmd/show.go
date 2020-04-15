package cmd

import (
	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
	"os"
)

var (
	asJSON bool
	asYAML bool
	asTree bool
)

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show nostromo configuration",
	Long: `Prints nostromo config with command tree
and profile changes.

By default the config file is located at ~/.nostromo/manifest.yaml.

Customize this with the $NOSTROMO_HOME environment variable`,
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(task.ShowConfig(asJSON, asYAML, asTree))
	},
}

func init() {
	manifestCmd.AddCommand(showCmd)

	showCmd.Flags().BoolVarP(&asJSON, "json", "j", false, "Show manifest as json")
	showCmd.Flags().BoolVarP(&asYAML, "yaml", "y", false, "Show manifest as yaml")
	showCmd.Flags().BoolVarP(&asTree, "tree", "t", false, "Show manifest as tree")
}
