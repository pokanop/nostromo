package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/pokanop/nostromo/shell"
	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

var (
	description string
	code        string
	language    string
	aliasOnly   bool
	mode        string
)

// addcmdCmd represents the addcmd command
var addcmdCmd = &cobra.Command{
	Use:   "cmd [key.path] [command] [options]",
	Short: "Add a command to nostromo manifest",
	Long: `Add a command to nostromo manifest for a given key path.
A key path is a '.' delimited string, e.g., "key.path" which represents
the alias which can be run as "key path" for the actual command provided.

This will create appropriate command scopes for all levels in
the provided key path. A command scope can a tree of sub commands
and substitutions.

A command's mode indicates how it will be executed. By default, nostromo
concatenates parent and child commands along the tree. There are 3 modes
available to commands:

  concatenate  Concatenate this command with subcommands exactly as defined
  independent  Execute this command with subcommands using ';' to separate
  exclusive    Execute this and only this command ignoring parent commands

You can set using -m or --mode when adding a command or globally using:
  nostromo manifest set mode <mode>`,
	Args: addCmdArgs,
	Run: func(cmd *cobra.Command, args []string) {
		var name string
		if len(args) > 1 {
			name = args[1]
		}
		os.Exit(task.AddCommand(args[0], name, description, code, language, aliasOnly, mode))
	},
}

func init() {
	addCmd.AddCommand(addcmdCmd)

	// Flags
	addcmdCmd.Flags().StringVarP(&description, "description", "d", "", "Description of the command to add")
	addcmdCmd.Flags().StringVarP(&code, "code", "c", "", "Code snippet to run for this command")
	addcmdCmd.Flags().StringVarP(&language, "language", "l", "", "Language of code snippet (e.g., ruby, python, perl, js)")
	addcmdCmd.Flags().BoolVarP(&aliasOnly, "alias-only", "a", false, "Add shell alias only, not a nostromo command")
	addcmdCmd.Flags().StringVarP(&mode, "mode", "m", "", "Set the mode for the command (concatenate, independent, exclusive)")
}

func codeValid() bool {
	return len(code) > 0 && len(language) > 0
}

func addCmdArgs(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("invalid number of arguments")
	}
	if len(args) < 2 && !codeValid() {
		return fmt.Errorf("must provide command or code snippet")
	}
	if codeValid() && !shell.IsSupportedLanguage(language) {
		return fmt.Errorf("invalid code snippet and language, must be in [%s]", strings.Join(shell.SupportedLanguages(), ","))
	}
	return nil
}
