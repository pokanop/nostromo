package cmd

import (
	"os"

	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

// docsCmd represents the hidden docs command used to generate the command
// reference for the documentation site.
var docsCmd = &cobra.Command{
	Use:   "docs [dir]",
	Short: "Generate markdown command reference",
	Long: `Generate markdown command reference for the documentation site.

Writes one markdown page per command into the given directory along with a
SUMMARY.md navigation file. Defaults to docs/reference when no directory is
provided. Existing generated pages in the directory are replaced.`,
	Hidden: true,
	Args:   cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dir := "docs/reference"
		if len(args) > 0 {
			dir = args[0]
		}
		os.Exit(task.GenerateDocs(cmd.Root(), dir))
	},
}

func init() {
	rootCmd.AddCommand(docsCmd)
}
