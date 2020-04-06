package cmd

import (
	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add commands or substitutions to nostromo",
	Long:  "Add commands or substitutions to nostromo",
	Run: func(cmd *cobra.Command, args []string) {
		task.AddInteractive()
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
