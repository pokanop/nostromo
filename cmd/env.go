package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/pokanop/nostromo/shell"
	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

var envShell string

// envCmd represents the env command
var envCmd = &cobra.Command{
	Use:   "env [key.path]",
	Short: "Print the effective environment for a command",
	Long: `Print the effective environment for a command in nostromo.

Merges the env vars and .env files of the command at "key.path" with those
inherited from its parents, in the order they are exported when the command
runs. Vars print as KEY=VALUE lines, use -v to see where each was defined or
-s to print export statements for a shell:
  nostromo env foo.bar
  nostromo env foo.bar -v
  nostromo env foo.bar -s fish`,
	Args: envCmdArgs,
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(task.Env(args[0], envShell))
	},
}

func envCmdArgs(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("invalid number of arguments")
	}
	if len(envShell) > 0 && !shell.IsSupported(envShell) {
		return fmt.Errorf("unsupported shell %q, must be in [%s]", envShell, strings.Join(shell.Shells, ","))
	}
	return nil
}

func init() {
	rootCmd.AddCommand(envCmd)

	// Flags
	envCmd.Flags().StringVarP(&envShell, "shell", "s", "", "Print export statements for shell (bash, zsh, fish, powershell)")
}
