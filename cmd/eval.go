package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/pokanop/nostromo/log"
	"github.com/pokanop/nostromo/model"
	"github.com/pokanop/nostromo/shell"
	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
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
	Args: func(cmd *cobra.Command, args []string) error {
		evalArgs, _, err := evalFlags(args)
		if err != nil {
			return err
		}
		return cobra.MinimumNArgs(1)(cmd, evalArgs)
	},
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		evalArgs, opts, err := evalFlags(args)
		if err != nil {
			log.Error(err)
			os.Exit(-1)
		}
		if opts.verbose {
			log.SetVerbose(true)
			model.SetVerbose(true)
		}
		os.Exit(task.EvalString(opts.shell, evalArgs))
	},
}

type evalOptions struct {
	verbose bool
	shell   string
}

// evalFlags consumes leading -v/--verbose and --shell flags that cobra cannot
// parse since flag parsing is disabled to pass user arguments through
// untouched. The shell is used to render env exports and defaults to bash.
func evalFlags(args []string) ([]string, evalOptions, error) {
	opts := evalOptions{shell: shell.Bash}
	for len(args) > 0 {
		if v, ok := verboseFlagValue(args[0]); ok {
			opts.verbose = v
			args = args[1:]
			continue
		}
		if sh, n, ok := shellFlagValue(args); ok {
			if !shell.IsSupported(sh) {
				return nil, opts, fmt.Errorf("unsupported shell %q, must be one of %s", sh, strings.Join(shell.Shells, ", "))
			}
			opts.shell = sh
			args = args[n:]
			continue
		}
		break
	}
	return args, opts, nil
}

// shellFlagValue returns the shell from a leading --shell flag along with the
// number of arguments consumed
func shellFlagValue(args []string) (string, int, bool) {
	switch {
	case args[0] == "--shell" || args[0] == "-s":
		if len(args) > 1 {
			return args[1], 2, true
		}
		return "", 1, true
	case strings.HasPrefix(args[0], "--shell="):
		return strings.TrimPrefix(args[0], "--shell="), 1, true
	case strings.HasPrefix(args[0], "-s="):
		return strings.TrimPrefix(args[0], "-s="), 1, true
	}
	return "", 0, false
}

func verboseFlagValue(arg string) (bool, bool) {
	switch {
	case arg == "-v" || arg == "--verbose":
		return true, true
	case strings.HasPrefix(arg, "-v=") || strings.HasPrefix(arg, "--verbose="):
		return strings.HasSuffix(arg, "=true"), true
	}
	return false, false
}

func init() {
	rootCmd.AddCommand(evalCmd)
}
