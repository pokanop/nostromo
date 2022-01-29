package cmd

import (
	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:    "run",
	Short:  "Placeholder node for nostromo commands",
	Long:   `Placeholder node for nostromo commands`,
	Hidden: true,
	Run:    func(cmd *cobra.Command, args []string) {},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Add all nostromo commands to the run command
	runCmd.AddCommand(task.FetchCommands()...)
}
