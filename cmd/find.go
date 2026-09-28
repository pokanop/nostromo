package cmd

import (
	"os"

	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

var (
	findExact bool
	findAll   bool
)

// findCmd represents the find command
var findCmd = &cobra.Command{
	Use:   "find [name|key.path]",
	Short: "Find commands and substitutions by name or key path",
	Long: `Find matching commands and substitutions in nostromo.

If the argument is the exact key path of a command in any docked manifest,
only that command and its substitutions are printed:

	nostromo find dev.install.goenv

Otherwise every command whose name, alias or key path contains the argument
is printed along with any matching substitutions:

	nostromo find install

Use --verbose to print results as tables including the manifest each match
belongs to. Use --exact to require an exact key path match and --all to list
every match even when the argument is itself a key path.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(task.Find(args[0], findExact, findAll))
	},
}

func init() {
	rootCmd.AddCommand(findCmd)

	// Flags
	findCmd.Flags().BoolVarP(&findExact, "exact", "e", false, "Only match a command at exactly the given key path")
	findCmd.Flags().BoolVarP(&findAll, "all", "a", false, "List all matches even when the argument is an exact key path")
}
