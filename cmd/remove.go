package cmd

import (
	"github.com/spf13/cobra"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove commands or substitutions from nostromo",
	Long:  "Remove commands or substitutions from nostromo",
	Run: func(cmd *cobra.Command, args []string) {
		printUsage(cmd)
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
