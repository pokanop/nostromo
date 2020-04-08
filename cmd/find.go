package cmd

import (
	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
	"os"
)

// findCmd represents the find command
var findCmd = &cobra.Command{
	Use:   "find [name]",
	Short: "Find matching commands and substitutions",
	Long: `Find matching commands and substitutions in nostromo.

Searches for "name" in commands and substitutions and prints matches.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(task.Find(args[0]))
	},
}

func init() {
	rootCmd.AddCommand(findCmd)
}
