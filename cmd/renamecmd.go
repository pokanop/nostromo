package cmd

import (
	"os"

	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

// renamecmdCmd represents the movecmd command
var renamecmdCmd = &cobra.Command{
	Use:   "cmd [key.path] [name] [options]",
	Short: "Rename a command in nostromo manifest",
	Long: `Rename a command in nostromo manifest for a given key path.
A key path is a '.' delimited string, e.g., "key.path" which represents
the alias which can be run as "key path" for the actual command provided.

Manipulate the core nostromo manifest by renaming command nodes.
Using rename, take a source key path and rename it to a new name with
all subcommands in tow.

If the new name already exists, the command will fail.

Optionally provide a description with -d to the destination.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(task.RenameCommand(args[0], args[1], description))
	},
}

func init() {
	renameCmd.AddCommand(renamecmdCmd)

	renamecmdCmd.Flags().StringVarP(&description, "description", "d", "", "Description of the command to rename")
}
