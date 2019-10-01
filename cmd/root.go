package cmd

import (
	"os"

	"github.com/pokanop/nostromo/log"
	"github.com/pokanop/nostromo/task"
	"github.com/pokanop/nostromo/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ver *version.Info

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "nostromo",
	Short: "Nostromo is a tool to manage aliases",
	Long: `Nostromo is a CLI to manage aliases through simple commands to add and remove
scoped aliases and substitutions.

Managing aliases can be tedius and difficult to set up. Nostromo makes this process easy
and reliable. The tool adds shortcuts to your .bashrc that call into the nostromo binary.
Nostromo reads and manages all aliases within its own config file.
This is used to find and execute the actual command intended as well as any
substitutions to simplify calls.`,
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
	ver = &version.Info{
		SemVer:    v,
		GitCommit: c,
		BuildDate: d,
	}

	// Update task package
	task.SetVersion(ver)
}

func init() {
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigName("env")             // name of config file (without extension)
	viper.AddConfigPath("$HOME")           // adding home directory as first search path
	viper.AddConfigPath("$HOME/.nostromo") // adding .nostromo directory to search path
	viper.AutomaticEnv()                   // read in environment variables that match
}

func printUsage(cmd *cobra.Command) {
	log.Regular(cmd.Long)
	log.Regular()
	log.Regular(cmd.UsageString())
}

func printVersion() {
	log.Regularf("Nostromo: %s\n", ver.Formatted())
}
