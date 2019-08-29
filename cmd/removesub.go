package cmd

import (
	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

// removesubCmd represents the removesub command
var removesubCmd = &cobra.Command{
	Use:   "sub [key.path] [alias]",
	Short: "Remove a substitution from nostromo manifest",
	Long: `Remove a substitution from nostromo manifest for a given key path and arg.
A substitution allows any arguments as part of a command to be substituted
by that value, e.g., "original-cmd //some/long/arg1 //some/long/arg2" can
be substituted out with "cmd-alias sub1 sub2" by adding subs for the args.

This will remove the substitution for scopes beneath levels in
the provided key path. A command scope can a tree of sub commands
and substitutions.`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		task.RemoveSubstitution(args[0], args[1])
	},
}

func init() {
	removeCmd.AddCommand(removesubCmd)
}
