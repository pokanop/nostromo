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
	Short: "Generates bash completion scripts",
	Long: `To load completion now, run

eval "$(nostromo completion)"

To configure your bash shell to load completions for each session add to your shell init files

# In ~/.bashrc or ~/.bash_profile
eval "$(nostromo completion)"`,
	Run: func(cmd *cobra.Command, args []string) {
		if writeCompletion {
			err := rootCmd.GenBashCompletionFile(pathutil.Abs("~/.nostromo/completion"))
			if err != nil {
				log.Error(err)
			} else {
				log.Highlight("bash completion script written to ~/.nostromo/completion")
			}
		} else {
			err := rootCmd.GenBashCompletion(os.Stdout)
			if err != nil {
				log.Error(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)

	completionCmd.Flags().BoolVarP(&writeCompletion, "write", "w", false, "Write bash completion to file")
}
