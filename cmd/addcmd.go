package cmd

import (
	"fmt"
	"strings"

	"github.com/pokanop/nostromo/shell"
	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

var (
	description string
	code        string
	language    string
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
and substitutions.`,
	Args: addCmdArgs,
	Run: func(cmd *cobra.Command, args []string) {
		name := ""
		if len(args) > 1 {
			name = args[1]
		}
		task.AddCommand(args[0], name, description, code, language)
	},
}

func init() {
	addCmd.AddCommand(addcmdCmd)

	// Flags
	addcmdCmd.Flags().StringVarP(&description, "description", "d", "", "Description of the command to add")
	addcmdCmd.Flags().StringVarP(&code, "code", "c", "", "Code snippet to run for this command")
	addcmdCmd.Flags().StringVarP(&language, "language", "l", "", "Language of code snippet (e.g., ruby, python, perl, js)")
}

func codeEmpty() bool {
	return len(code) == 0 && len(language) == 0
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
		return fmt.Errorf("invalid code snippet and language, must be in [%s]", strings.Join(shell.ValidLanguages(), ","))
	}
	return nil
}
