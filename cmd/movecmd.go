package cmd

import (
	"os"

	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

var rename string

// movecmdCmd represents the movecmd command
var movecmdCmd = &cobra.Command{
	Use:   "cmd [src.key.path] [dest.key.path] [options]",
	Short: "Move a command in nostromo manifest",
	Long: `Move a command in nostromo manifest for a given key path.
A key path is a '.' delimited string, e.g., "key.path" which represents
the alias which can be run as "key path" for the actual command provided.

Manipulate the core nostromo manifest by moving command nodes.
Using move, take a source key path and shift it to the destination
key path with all subcommands in tow.

If the destination key path does not exist, it will be created
and the node will be moved there. Optionally provide a description
with -d to the destination.

To move a node to the root, do not provide a destination key path.

To rename a node only, pass the -r flag with a name to use.`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		var dest string
		if len(args) > 1 {
			dest = args[1]
		}
		if len(rename) == 0 {
			os.Exit(task.MoveCommand(args[0], dest, description))
		} else {
			os.Exit(task.RenameCommand(args[0], rename, description))
		}
	},
}

func init() {
	moveCmd.AddCommand(movecmdCmd)

	movecmdCmd.Flags().StringVarP(&description, "description", "d", "", "Description of the destination to move command")
	movecmdCmd.Flags().StringVarP(&rename, "rename", "r", "", "Rename command instead of moving")
}
