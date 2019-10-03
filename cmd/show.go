package cmd

import (
	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

var (
	rawJSON bool
	rawYAML bool
)

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show nostromo configuration",
	Long: `Prints nostromo config with command tree
and profile changes.

The config file is located at ~/.nostromo/config`,
	Run: func(cmd *cobra.Command, args []string) {
		task.ShowConfig(rawJSON, rawYAML)
	},
}

func init() {
	manifestCmd.AddCommand(showCmd)

	showCmd.Flags().BoolVarP(&rawJSON, "json", "j", false, "Show manifest as json")
	showCmd.Flags().BoolVarP(&rawYAML, "yaml", "y", false, "Show manifest as yaml")
}
