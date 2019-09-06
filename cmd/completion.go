package cmd

import (
	"os"

	"github.com/pokanop/nostromo/log"
	"github.com/pokanop/nostromo/pathutil"
	"github.com/spf13/cobra"
)

var writeCompletion bool

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generates shell completion scripts",
	Long: `To load completion now, run

eval "$(nostromo completion)"	# for bash
eval "$(nostromo completion --zsh)" # for zsh

To configure your shell to load completions for each session add to your init files

# In ~/.bashrc or ~/.bash_profile
eval "$(nostromo completion)"

# In ~/.zshrc
eval "$(nostromo completion --zsh)"`,
	Run: func(cmd *cobra.Command, args []string) {
		if writeCompletion {
			writeCompletionFile()
		} else {
			printCompletion()
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)

	completionCmd.Flags().BoolVarP(&writeCompletion, "write", "w", false, "Write completions to file")
	completionCmd.Flags().BoolVarP(&useZsh, "zsh", "z", false, "Generate completions for zsh")
}

func writeCompletionFile() {
	if useZsh {
		writeZshCompletionFile()
	} else {
		writeBashCompletionFile()
	}
}

func writeBashCompletionFile() {
	err := rootCmd.GenBashCompletionFile(pathutil.Abs("~/.nostromo/completion"))
	if err != nil {
		log.Error(err)
	} else {
		log.Highlight("bash completion script written to ~/.nostromo/completion")
	}
}

func writeZshCompletionFile() {
	err := rootCmd.GenZshCompletionFile(pathutil.Abs("~/.nostromo/zcompletion"))
	if err != nil {
		log.Error(err)
	} else {
		log.Highlight("zsh completion script written to ~/.nostromo/zcompletion")
	}
}

func printCompletion() {
	if useZsh {
		printZshCompletion()
	} else {
		printBashCompletion()
	}
}

func printBashCompletion() {
	err := rootCmd.GenBashCompletion(os.Stdout)
	if err != nil {
		log.Error(err)
	}
}

func printZshCompletion() {
	err := rootCmd.GenZshCompletion(os.Stdout)
	if err != nil {
		log.Error(err)
	}
}
