package cmd

import (
	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
	"os"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a config item from manifest",
	Long: `Get a config item from manifest.
Nostromo config items are saved in the manifest.

Use this command to get keys to examine these settings:
verbose: true | false
aliasesOnly: true | false`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(task.GetConfig(args[0]))
	},
}

func init() {
	manifestCmd.AddCommand(getCmd)
}
