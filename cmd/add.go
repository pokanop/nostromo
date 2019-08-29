package cmd

import (
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add commands or substitutions to nostromo",
	Long:  "Add commands or substitutions to nostromo",
	Run: func(cmd *cobra.Command, args []string) {
		printUsage(cmd)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
