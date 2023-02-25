package cmd

import (
	"os"

	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

var force bool
var keep bool

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync [name]...",
	Short: "Sync docked manifests from source locations",
	Long: `Sync docked manifests from source locations and
makes commands available for execution.

Sync can be used to update previously docked manifests from
respective data sources. Provide one or more of the names of 
the manifests as arguments to sync.

Sync will only update manifests with changed identifiers, to
force update use the -f flag.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(task.Sync(force, keep, args))
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	syncCmd.Flags().BoolVarP(&force, "force", "f", false, "Force sync manifests")
	syncCmd.Flags().BoolVarP(&keep, "keep", "k", false, "Keep downloaded files")
}
