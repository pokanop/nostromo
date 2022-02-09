package cmd

import (
	"os"

	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

// copycmdCmd represents the copycmd command
var copycmdCmd = &cobra.Command{
	Use:   "cmd [src.key.path] [dest.key.path] [options]",
	Short: "Copy a command in nostromo manifest",
	Long: `Copy a command in nostromo manifest for a given key path.
A key path is a '.' delimited string, e.g., "key.path" which represents
the alias which can be run as "key path" for the actual command provided.

Manipulate nostromo manifests by copying command nodes. Using copy, take
a source key path and copy it to the destination key path with all 
subcommands in tow.

If the destination key path does not exist, it will be created
and the node will be copied there. Optionally provide a description
with -d to the destination.

To copy a node to the root, use '.' for destination. Optionally provide
a destination manifest using the -m flag. Otherwise the copy will default 
to the core manifest.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(task.MoveCommand(args[0], args[1], manifest, description, true))
	},
}

func init() {
	copyCmd.AddCommand(copycmdCmd)

	copycmdCmd.Flags().StringVarP(&description, "description", "d", "", "Description of the destination to copy command to")
	copycmdCmd.Flags().StringVarP(&manifest, "manifest", "m", "", "Destination manifest to copy command to")
}
