package cmd

import (
	"os"

	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update [key.path] [command] [options]",
	Short: "Update a command in nostromo manifest",
	Long: `Update a command in nostromo manifest for a given key path.
A key path is a '.' delimited string, e.g., "key.path" which represents
the alias which can be run as "key path" for the actual command provided.

This will update appropriate command scopes for all levels in the provided
key path. A command scope can contain a tree of sub commands and 
substitutions.

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
		os.Exit(task.AddCommand(args[0], name, description, code, language, aliasOnly, mode, true))
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	// Flags
	updateCmd.Flags().StringVarP(&description, "description", "d", "", "Description of the command to update")
	updateCmd.Flags().StringVarP(&code, "code", "c", "", "Code snippet to run for this command")
	updateCmd.Flags().StringVarP(&language, "language", "l", "", "Language of code snippet (e.g., ruby, python, perl, js)")
	updateCmd.Flags().BoolVarP(&aliasOnly, "alias-only", "a", false, "Add shell alias only, not a nostromo command")
	updateCmd.Flags().StringVarP(&mode, "mode", "m", "", "Set the mode for the command (concatenate, independent, exclusive)")
}
