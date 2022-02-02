package cmd

import (
	"os"

	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

var force bool

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync docked manifests from source locations",
	Long: `Sync docked manifests from source locations and
makes commands available for execution.

Sync can be used to update previously docked manifests from
respective data sources.

Sync will only update manifests with changed identifiers, to
force update use the -f flag.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(task.Sync(force, []string{}))
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	syncCmd.Flags().BoolVarP(&force, "force", "f", false, "Force sync manifests")
}
