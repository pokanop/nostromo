package cmd

import (
	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

var useZsh bool

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run [command] [args]",
	Short: "Run a command from manifest",
	Long: `Run a command from manifest.
After adding commands you can run them through nostromo. As long as
a command can be found in the manifest it will be executed.

If you create key paths to commands like this:
  nostromo add cmd foo.bar "./crazy-long-command with-args"
  nostromo add cmd foo.baz "./another-crazy-long-command with-args"
  
It will create command entries at each level. Now you can run the following
it will execute the actual commands:
  nostromo run foo bar
  nostromo run foo baz
  
Since nostromo will alias these commands in your shell profile you can omit
the "nostromo run" command itself. For example, it will add:
  alias foo='nostromo run foo "$*"'

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
		task.Run(args, useZsh)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().BoolVarP(&useZsh, "zsh", "z", false, "Use zsh for shell exec")
}
