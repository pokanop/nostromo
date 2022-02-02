package cmd

import (
	"os"

	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

var nuke bool

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy nostromo configuration",
	Long: `Destroy nostromo configuration and start fresh.

By default the core manifest is only destroyed and recreated.

Optionally delete the entire installation using -n flag. Note that
this does not remove shell init file entries added by nostromo. 
You'll have to delete those manually.`,
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(task.DestroyConfig(nuke))
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)

	destroyCmd.Flags().BoolVarP(&nuke, "nuke", "n", false, "Nuke the entire installation")
}
