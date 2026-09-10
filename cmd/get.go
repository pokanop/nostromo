package cmd

import (
	"os"

	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a config item from the spaceport",
	Long: `Get a config item from the spaceport.
Nostromo config items are global and saved in the spaceport.

Use this command to get keys to examine these settings:
  verbose: boolean
  aliasesOnly: boolean
  mode: concatenate | independent | exclusive
  backupCount: number
  theme: default | grayscale | emoji`,
	Args:      cobra.MinimumNArgs(1),
	ValidArgs: []string{"verbose", "aliasesOnly", "mode", "backupCount", "theme"},
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(task.GetConfig(args[0]))
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
