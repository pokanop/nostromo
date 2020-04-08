package cmd

import (
	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
	"os"
)

// evalCmd represents the eval command
var evalCmd = &cobra.Command{
	Use:   "eval [command] [args]",
	Short: "Show eval command from manifest",
	Long: `Show eval command from manifest.
After adding commands you can run them through nostromo. As long as
a command can be found in the manifest it will provide a command to eval.

If you create key paths to commands like this:
  nostromo add cmd foo.bar "./crazy-long-command with-args"
  nostromo add cmd foo.baz "./another-crazy-long-command with-args"
  
It will create command entries at each level. Nostromo will alias these 
commands in your shell profile to be evaluated. 
For example, it will add:
  alias foo=eval 'nostromo eval foo "$*"'

The power of these commands are that you can take complicated combinations
of commands and build up a more intelligent way to run things as if you wrote
your own tool. Imagine composing commands to simplify a workflow:
  build thing1
  build thing2

The root "build" command can do things like cd to a folder, set env vars, and
run the main command. Lastly, substitutions can further shorten any sets of
commands that need to be run across the scope of the command.`,
	Args:               cobra.MinimumNArgs(1),
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(task.EvalString(args))
	},
}

func init() {
	rootCmd.AddCommand(evalCmd)
}
