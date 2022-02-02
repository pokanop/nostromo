package cmd

import (
	"os"

	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

// moveCmd represents the move command
var moveCmd = &cobra.Command{
	Use:   "move [source] [dest] [options]",
	Short: "Move a command from one key path to another",
	Long: `Move a command from one key path to another.

Manipulate the core nostromo manifest by moving command nodes.
Using move, take a source key path and shift it to the destination
key path with all subcommands in tow.

If the destination key path does not exist, it will be created
and the node will be moved there. Optionally provide a description
with -d to the destination.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(task.MoveCommand(args[0], args[1], description))
	},
}

func init() {
	rootCmd.AddCommand(moveCmd)

	moveCmd.Flags().StringVarP(&description, "description", "d", "", "Description of the command to add")
}
