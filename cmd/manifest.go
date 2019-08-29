package cmd

import (
	"github.com/spf13/cobra"
)

// manifestCmd represents the manifest command
var manifestCmd = &cobra.Command{
	Use:   "manifest",
	Short: "Interact with nostromo manifest",
	Long:  "Interact with nostromo manifest",
	Run: func(cmd *cobra.Command, args []string) {
		printUsage(cmd)
	},
}

func init() {
	rootCmd.AddCommand(manifestCmd)
}
