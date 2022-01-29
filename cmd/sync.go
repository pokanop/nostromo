package cmd

import (
	"os"

	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

var force bool

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync [source] [options]",
	Short: "Sync nostromo manifests from source locations",
	Long: `Sync nostromo manifests from source locations and
makes commands available for execution.

Sync can be used to copy a single manifest to nostromo's config
folder. Or it can be used to update previously synchronized 
manifests from respective data sources.

Providing a [source] file will sync only that file. Omitting it
directs nostromo to sync all manifests.`,
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(task.Sync(force, args))
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	syncCmd.Flags().BoolVarP(&force, "force", "f", false, "Force sync manifest(s)")
}
