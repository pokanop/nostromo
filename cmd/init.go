package cmd

import (
	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
	"os"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize nostromo config file",
	Long: `Create a nostromo config file with defaults.
	
The config file is located at ~/.nostromo/config`,
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(task.InitConfig())
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
