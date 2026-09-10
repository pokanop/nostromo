package cmd

import (
	"os"

	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

var linkTarget string

// linkCmd represents the link command
var linkCmd = &cobra.Command{
	Use:   "link [source] [options]",
	Short: "Link a manifest as a dependency",
	Long: `Link a manifest from a source location as a dependency of another
manifest, docking it along with any manifests it links itself.

The link is recorded on the target manifest, which is the core manifest
unless --to is given, so syncing the target fetches the linked manifest
as well. Linking fails if a manifest with the same name already exists
from a different source or if the link would be circular.

Run:

	nostromo link https://foo.com/edit.yaml
	nostromo link file://path/to/install.yaml --to tools

Use unlink to remove the dependency again.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(task.Link(args[0], linkTarget, force, keep))
	},
}

func init() {
	rootCmd.AddCommand(linkCmd)

	linkCmd.Flags().StringVarP(&linkTarget, "to", "t", "", "Manifest to record the link on (default: core manifest)")
	linkCmd.Flags().BoolVarP(&force, "force", "f", false, "Force update of the linked manifest")
	linkCmd.Flags().BoolVarP(&keep, "keep", "k", false, "Keep downloaded files")
}
