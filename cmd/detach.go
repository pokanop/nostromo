package cmd

import (
	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

var (
	detachKeyPath      string
	detachDescription  string
	detachKeepOriginal bool
)

// detachCmd represents the detach command
var detachCmd = &cobra.Command{
	Use:   "detach [name] [key.path]... [options]",
	Short: "Detach a command node into a new manifest",
	Long: `Detach will extract an entire command tree from the given node and
create a brand new manifest under the ships/ folder.

Running:

	nostromo detach mobile-builds build.ios build.android

creates a new mobile-builds.yaml file with build command sets joined in the
new manifest. If a file with the target name already exists, nostromo will
attempt to merge the manifests automatically.

By default, the detached node is removed from the manifest. If this node should
be kept in the original manifest, use the --keep flag.

This is a convenient way to slice your command sets and produce manifests
that can be shared via the sync command.`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		task.Detach(args[0], args[1:], detachKeyPath, detachDescription, detachKeepOriginal)
	},
}

func init() {
	rootCmd.AddCommand(detachCmd)

	// Flags
	detachCmd.Flags().StringVarP(&detachKeyPath, "root", "r", "", "Add detached commands to a key path, defaults to root")
	detachCmd.Flags().StringVarP(&detachDescription, "description", "d", "", "A description for the destination key path")
	detachCmd.Flags().BoolVarP(&detachKeepOriginal, "keep", "k", false, "Keep original command tree intact")
}
