package cmd

import (
	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

// removecmdCmd represents the removecmd command
var removecmdCmd = &cobra.Command{
	Use:   "cmd [key.path]",
	Short: "Remove a command from nostromo manifest",
	Long: `Remove a command to nostromo manifest for a given key path.
A key path is a '.' delimited string, e.g., "key.path" which represents
the alias which can be run as "key path" for the actual command provided.
	
This will remove appropriate command scopes for all levels beneath
the provided key path. A command scope can a tree of sub commands
and substitutions.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		task.RemoveCommand(args[0])
	},
}

func init() {
	removeCmd.AddCommand(removecmdCmd)
}
