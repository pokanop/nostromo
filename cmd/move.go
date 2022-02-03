package cmd

import (
	"github.com/spf13/cobra"
)

// moveCmd represents the move command
var moveCmd = &cobra.Command{
	Use:   "move",
	Short: "Move a command in nostromo manifest",
	Long:  "Move a command in nostromo manifest",
	Run: func(cmd *cobra.Command, args []string) {
		printUsage(cmd)
	},
}

func init() {
	rootCmd.AddCommand(moveCmd)
}
