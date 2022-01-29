package cmd

import (
	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Hidden: true,
	Run:    func(cmd *cobra.Command, args []string) {},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Add all nostromo commands to the run command
	runCmd.AddCommand(task.FetchCommands()...)
}
