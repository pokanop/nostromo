package cmd

import (
	"os"

	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

var manifest string

// movecmdCmd represents the movecmd command
var movecmdCmd = &cobra.Command{
	Use:   "cmd [src.key.path] [dest.key.path] [options]",
	Short: "Move a command in nostromo manifest",
	Long: `Move a command in nostromo manifest for a given key path.
A key path is a '.' delimited string, e.g., "key.path" which represents
the alias which can be run as "key path" for the actual command provided.

Manipulate nostromo manifests by moving command nodes. Using move, take
a source key path and shift it to the destination key path with all 
subcommands in tow.

If the destination key path does not exist, it will be created
and the node will be moved there. Optionally provide a description
with -d to the destination.

To move a node to the root, use '.' for destination. Optionally provide
a destination manifest using the -m flag. Otherwise the move will default 
to the core manifest.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(task.MoveCommand(args[0], args[1], manifest, description, false))
	},
}

func init() {
	moveCmd.AddCommand(movecmdCmd)

	movecmdCmd.Flags().StringVarP(&description, "description", "d", "", "Description of the destination to move command to")
	movecmdCmd.Flags().StringVarP(&manifest, "manifest", "m", "", "Destination manifest to move command to")
}
