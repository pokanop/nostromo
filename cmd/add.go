package cmd

import (
	"os"

	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add command or substitution",
	Long:  "Add command or substitution",
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(task.AddInteractive())
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
