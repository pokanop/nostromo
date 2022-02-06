package cmd

import (
	"os"

	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

// dockCmd represents the dock command
var dockCmd = &cobra.Command{
	Use:   "dock [source]... [options]",
	Short: "Dock a nostromo manifest",
	Long: `Dock a nostromo manifest from source location and
makes commands available for execution.

Dock can be used to copy a single manifest to nostromo's config
folder. If a docked manifest already exists with the same name then
nostromo will overwrite that file if the identifier is different.

To force docking even if identifiers are the same, use the -f flag.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(task.Sync(force, args))
	},
}

func init() {
	rootCmd.AddCommand(dockCmd)

	dockCmd.Flags().BoolVarP(&force, "force", "f", false, "Force dock manifest")
}
