package cmd

import (
	"os"

	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

// linksCmd represents the links command
var linksCmd = &cobra.Command{
	Use:   "links [name]",
	Short: "Show the manifest dependency graph",
	Long: `Show the dependency graph of linked manifests.

Without a name every docked manifest is shown along with the manifests
it links. Pass a manifest name to show only its dependencies. Circular
references and links whose manifest is not docked are marked.

Run:

	nostromo links
	nostromo links tools`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := ""
		if len(args) > 0 {
			name = args[0]
		}
		os.Exit(task.Links(name))
	},
}

func init() {
	rootCmd.AddCommand(linksCmd)
}
