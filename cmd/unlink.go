package cmd

import (
	"os"

	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

var unlinkTarget string

// unlinkCmd represents the unlink command
var unlinkCmd = &cobra.Command{
	Use:   "unlink [name] [options]",
	Short: "Unlink a manifest dependency",
	Long: `Unlink a manifest dependency from another manifest.

The link is removed from the target manifest, which is the core manifest
unless --to is given. The linked manifest is undocked when nothing else
links it and it was not docked directly, along with any manifests only
it pulled in.

Run:

	nostromo unlink edit
	nostromo unlink install --to tools`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(task.Unlink(args[0], unlinkTarget))
	},
}

func init() {
	rootCmd.AddCommand(unlinkCmd)

	unlinkCmd.Flags().StringVarP(&unlinkTarget, "to", "t", "", "Manifest to remove the link from (default: core manifest)")
}
