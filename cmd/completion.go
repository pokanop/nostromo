package cmd

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"

	"github.com/pokanop/nostromo/log"
	"github.com/pokanop/nostromo/pathutil"
	"github.com/spf13/cobra"
)

const (
	bashCompletionPath = "~/.nostromo/completion.bash"
	zshCompletionPath  = "~/.nostromo/completion.zsh"
)

var useZsh bool
var writeCompletion bool

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generates shell completion scripts",
	Long: `To load completion now, run

eval "$(nostromo completion)"	# for bash
eval "$(nostromo completion --zsh)" # for zsh

To configure your shell to load completions for each session add to your init files.
Note that "nostromo init" will add this automatically.

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
	err := rootCmd.GenBashCompletionFile(pathutil.Abs(bashCompletionPath))
	if err != nil {
		log.Error(err)
	} else {
		log.Highlight("bash completion script written to " + bashCompletionPath)
	}
}

func writeZshCompletionFile() {
	var buf bytes.Buffer
	err := rootCmd.GenZshCompletion(&buf)
	if err != nil {
		log.Error(err)
		return
	}

	f, err := os.Create(pathutil.Abs(zshCompletionPath))
	if err != nil {
		log.Error(err)
		return
	}

	fixZshCompletion(&buf, f)

	log.Highlight("zsh completion script written to " + zshCompletionPath)
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
	var buf bytes.Buffer
	err := rootCmd.GenZshCompletion(&buf)
	if err != nil {
		log.Error(err)
		return
	}
	fixZshCompletion(&buf, os.Stdout)
}

// This is required due to a bug in cobra and zsh support:
// https://github.com/spf13/cobra/pull/887
func fixZshCompletion(r io.Reader, w io.Writer) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		log.Error(err)
		return
	}
	s := string(b)
	s += "compdef _nostromo nostromo\n"
	w.Write([]byte(s))
}
