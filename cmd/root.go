package cmd

import (
	"os"

	"github.com/pokanop/nostromo/config"
	"github.com/pokanop/nostromo/log"
	"github.com/pokanop/nostromo/model"
	"github.com/pokanop/nostromo/task"
	"github.com/pokanop/nostromo/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ver *version.Info
var verbose bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "nostromo",
	Short: "nostromo is a tool to manage aliases",
	Long: `nostromo is a CLI to manage aliases through simple commands to add and remove
scoped aliases and substitutions.

Managing aliases can be tedious and difficult to set up. nostromo makes this process easy
and reliable. The tool adds shortcuts to your .bashrc that call into the nostromo binary.
nostromo reads and manages all aliases within its own config file.
This is used to find and execute the actual command intended as well as any
substitutions to simplify calls.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.SetVerbose(verbose)
		model.SetVerbose(verbose)
	},
	SilenceErrors: true,
	SilenceUsage:  true,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(-1)
	}
}

// SetVersion to inject version info
func SetVersion(v, c, d string) {
	ver = version.NewInfo(v, c, d)

	// Update dependent packages
	task.SetVersion(ver)
	config.SetVersion(ver)
}

func init() {
	cobra.OnInitialize(initConfig)

	// Disable default completion command
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "show verbose logging")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	nostromoPath := config.BaseDir()  // get base path. (by default ~/.nostromo)
	viper.SetConfigName("env")        // name of config file (without extension)
	viper.AddConfigPath("$HOME")      // adding home directory as first search path
	viper.AddConfigPath(nostromoPath) // adding nostromo directory to search path
	viper.AutomaticEnv()              // read in environment variables that match
}

func printUsage(cmd *cobra.Command) {
	log.Regular(cmd.Long)
	log.Regular()
	log.Regular(cmd.UsageString())
}

func printVersion() {
	log.Regularf("nostromo: %s\n", ver.Formatted())
}
