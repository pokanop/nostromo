package cmd

import (
	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
	"os"
)

// addsubCmd represents the addsub command
var addsubCmd = &cobra.Command{
	Use:   "sub [key.path] [name] [alias]",
	Short: "Add a substitution to nostromo manifest",
	Long: `Add a substitution to nostromo manifest for a given key path and arg.
A substitution allows any arguments as part of a command to be substituted
by that value, e.g., "original-cmd //some/long/arg1 //some/long/arg2" can
be substituted out with "cmd-alias sub1 sub2" by adding subs for the args.

This will create the substitution for scopes beneath levels in
the provided key path. A command scope can a tree of sub commands
and substitutions.`,
	Args: cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(task.AddSubstitution(args[0], args[1], args[2]))
	},
}

func init() {
	addCmd.AddCommand(addsubCmd)
}
