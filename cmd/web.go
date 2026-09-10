package cmd

import (
	"os"

	"github.com/pokanop/nostromo/web"
	"github.com/spf13/cobra"
)

var (
	webPort   int
	webNoOpen bool
)

// webCmd represents the web command
var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Serve a local web UI for editing manifests",
	Long: `Serve a local web UI for editing manifests.

Starts a web server bound to 127.0.0.1 and opens the UI in your browser.
The UI lists docked manifests, shows the command tree and lets you add,
edit, search and remove commands and substitutions in the core manifest.

Changes are saved through the same path as the CLI so cargo backups keep
working. Docked manifests are shown read-only since syncing overwrites them.

Press ctrl+c to stop the server.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(web.Serve(webPort, !webNoOpen, ver))
	},
}

func init() {
	rootCmd.AddCommand(webCmd)

	webCmd.Flags().IntVarP(&webPort, "port", "p", 8080, "Port to listen on, use 0 for a random free port")
	webCmd.Flags().BoolVar(&webNoOpen, "no-open", false, "Do not open the browser automatically")
}
