package cmd

import (
	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show nostromo configuration",
	Long: `Prints nostromo config as JSON.

The config file is located at ~/.nostromo/config`,
	Run: func(cmd *cobra.Command, args []string) {
		task.ShowConfig()
	},
}

func init() {
	manifestCmd.AddCommand(showCmd)
}
