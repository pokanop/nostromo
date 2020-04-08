package cmd

import (
	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
	"os"
)

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Delete nostromo configuration",
	Long: `Deletes nostromo config file.
	
The config file is located at ~/.nostromo/config`,
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(task.DestroyConfig())
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
}
