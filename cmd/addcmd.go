package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/pokanop/nostromo/dotenv"
	"github.com/pokanop/nostromo/model"
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
	platforms   []string
	env         []string
	unsetEnv    []string
	dotenvFiles []string
)

// addcmdCmd represents the addcmd command
var addcmdCmd = &cobra.Command{
	Use:   "cmd [key.path] [command] [options]",
	Short: "Add a command to nostromo manifest",
	Long: `Add a command to nostromo manifest for a given key path.
A key path is a '.' delimited string, e.g., "key.path" which represents
the alias which can be run as "key path" for the actual command provided.

This will create appropriate command scopes for all levels in the provided
key path. A command scope can contain a tree of sub commands and 
substitutions.

A command's mode indicates how it will be executed. By default, nostromo
concatenates parent and child commands along the tree. There are 3 modes
available to commands:

  concatenate  Concatenate this command with subcommands exactly as defined
  independent  Execute this command with subcommands using ';' to separate
  exclusive    Execute this and only this command ignoring parent commands

You can set using -m or --mode when adding a command or globally using:
  nostromo manifest set mode <mode>

A command can be limited to specific platforms with -p or --platforms using
Go OS names (e.g., linux, darwin, windows) or OS/arch pairs (e.g., linux/arm64).
Commands unavailable on the current platform, including their sub commands,
are not aliased or completed in the shell and cannot be run:
  nostromo add cmd foo.bar "pbcopy" --platforms darwin

Environment variables can be exported before a command runs with -e or --env
and loaded from .env files with --dotenv. Both are inherited by sub commands
which may override them, and values may reference other variables:
  nostromo add cmd foo "cd ~/foo" --dotenv ~/foo/.env --env APP_ENV=dev
  nostromo add cmd foo.bar "make" --env BIN='$PWD/bin:$PATH'

Use "nostromo env foo.bar" to see the effective environment for a command.`,
	Args: addCmdArgs,
	Run: func(cmd *cobra.Command, args []string) {
		var name string
		if len(args) > 1 {
			name = args[1]
		}
		os.Exit(task.AddCommand(args[0], name, description, code, language, aliasOnly, mode, platformsFlag(cmd), envFlags(cmd), false))
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
	addcmdCmd.Flags().StringSliceVarP(&platforms, "platforms", "p", nil, "Limit the command to platforms (e.g., linux,darwin,windows/arm64)")
	addEnvFlags(addcmdCmd)
}

// addEnvFlags shared by add and update
func addEnvFlags(cmd *cobra.Command) {
	cmd.Flags().StringArrayVarP(&env, "env", "e", nil, "Export an env var as KEY=VALUE before running, repeatable")
	cmd.Flags().StringArrayVar(&unsetEnv, "unset-env", nil, "Remove an env var by KEY, repeatable")
	cmd.Flags().StringArrayVar(&dotenvFiles, "dotenv", nil, "Load a .env file before running (~ and $VARS expanded), repeatable")
}

// envFlags returns the env changes from --env, --unset-env and --dotenv,
// dotenv is nil if the flag was not given so existing files are kept when
// updating and an empty --dotenv "" clears them
func envFlags(cmd *cobra.Command) model.EnvChanges {
	changes := model.EnvChanges{Unset: unsetEnv}
	changes.Set, _ = model.ParseEnv(env)
	if cmd.Flags().Changed("dotenv") {
		changes.Dotenv = []string{}
		for _, path := range dotenvFiles {
			if len(path) > 0 {
				changes.Dotenv = append(changes.Dotenv, path)
			}
		}
	}
	return changes
}

func envValid() error {
	if _, err := model.ParseEnv(env); err != nil {
		return err
	}
	for _, key := range unsetEnv {
		if !dotenv.ValidKey(key) {
			return fmt.Errorf("invalid env key %q", key)
		}
	}
	return nil
}

func codeValid() bool {
	return len(code) > 0 && len(language) > 0
}

// platformsFlag returns the parsed --platforms list, or nil if the flag was
// not given so existing platforms are kept when updating
func platformsFlag(cmd *cobra.Command) []string {
	if !cmd.Flags().Changed("platforms") {
		return nil
	}
	parsed, err := model.ParsePlatforms(platforms)
	if err != nil {
		return nil
	}
	return parsed
}

func platformsValid() error {
	_, err := model.ParsePlatforms(platforms)
	return err
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
	if err := envValid(); err != nil {
		return err
	}
	return platformsValid()
}
