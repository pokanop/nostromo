package cmd

import (
	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

var versionBanner bool

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version of nostromo",
	Long: `Print version of nostromo

Supplies tag version, commit hash, and date`,
	Run: func(cmd *cobra.Command, args []string) {
		if versionBanner {
			task.PrintBanner()
		}
		printVersion()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	versionCmd.Flags().BoolVar(&versionBanner, "banner", false, "print the nostromo ascii-art banner")
}
