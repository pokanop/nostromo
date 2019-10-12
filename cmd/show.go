package cmd

import (
	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
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

The config file is located at ~/.nostromo/config`,
	Run: func(cmd *cobra.Command, args []string) {
		task.ShowConfig(asJSON, asYAML, asTree)
	},
}

func init() {
	manifestCmd.AddCommand(showCmd)

	showCmd.Flags().BoolVarP(&asJSON, "json", "j", false, "Show manifest as json")
	showCmd.Flags().BoolVarP(&asYAML, "yaml", "y", false, "Show manifest as yaml")
	showCmd.Flags().BoolVarP(&asTree, "tree", "t", false, "Show manifest as tree")
}
