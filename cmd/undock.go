package cmd

import (
	"os"

	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

// undockCmd represents the undock command
var undockCmd = &cobra.Command{
	Use:   "undock [name]...",
	Short: "Undock nostromo manifests",
	Long: `Undock nostromo manifests and remove commands from being
executable.

A manifest added to nostromo using the dock command can be undocked. 
This will delete the file from the local configuration and the commands
will no longer be available to run.

Run:

	nostromo undock edit install

To get the commands back you will need to dock the manifest again from 
the original source location.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(task.Undock(args))
	},
}

func init() {
	rootCmd.AddCommand(undockCmd)
}
