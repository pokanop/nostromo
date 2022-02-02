package cmd

import (
	"os"

	"github.com/pokanop/nostromo/task"
	"github.com/spf13/cobra"
)

// uuidgenCmd represents the uuidgen command
var uuidgenCmd = &cobra.Command{
	Use:   "uuidgen [name]",
	Short: "Generate a new unique id for a manifest",
	Long: `Generate a new unique id for a manifest.

nostromo uses a uuid to determine if a manifest is unique or not.
When using sync to get new manifests, nostromo will only apply the
changes if it detects a different identifier.

This command allows regenerating the uuid to allow for publishing
and pulling updated manifests. Note that using standard nostromo
commands automatically updates the identifier. This command can be
used if manual updates were made to a manifest.

Omitting the name of the manifest will apply to the core manifest.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var name string
		if len(args) > 0 {
			name = args[0]
		}
		os.Exit(task.RegenerateID(name))
	},
}

func init() {
	rootCmd.AddCommand(uuidgenCmd)
}
