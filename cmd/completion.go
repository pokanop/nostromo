package cmd

import (
	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
	"os"
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generates shell completion scripts",
	Long: `To load completion now, run

eval "$(nostromo completion)"

To configure your shell to load completions for each session add to your init files.
Note that "nostromo init" will add this automatically.

# In ~/.bashrc, ~/.bash_profile or ~/.zshrc
eval "$(nostromo completion)"`,
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(task.GenerateCompletions(cmd))
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
