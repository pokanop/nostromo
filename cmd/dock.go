package cmd

import (
	"os"

	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

// dockCmd represents the dock command
var dockCmd = &cobra.Command{
	Use:   "dock [source]... [options]",
	Short: "Dock nostromo manifests",
	Long: `Dock nostromo manifests from source locations and make commands
available for execution.


Dock can be used to copy a single manifest or more to nostromo's config
folder. If a docked manifest already exists with the same name then
nostromo will overwrite that file if the identifier is different.

Run:

	nostromo dock http://foo.com/edit.yaml file://path/to/install.yaml

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
