package task

import (
	"strings"

	"github.com/pokanop/nostromo/config"
	"github.com/pokanop/nostromo/log"
	"github.com/pokanop/nostromo/model"
	"github.com/pokanop/nostromo/pathutil"
	"github.com/pokanop/nostromo/prompt"
	"github.com/pokanop/nostromo/shell"
	"github.com/pokanop/nostromo/stringutil"
	"github.com/pokanop/nostromo/version"
	"github.com/shivamMg/ppds/tree"
	"github.com/spf13/cobra"
)

var ver *version.Info

// SetVersion should be called before any task to ensure manifest is updated
func SetVersion(v *version.Info) {
	ver = v
}

// InitConfig of nostromo config file if not already initialized
func InitConfig() int {
	cfg := checkConfigQuiet()

	if cfg == nil {
		baseDir := config.GetBaseDir()
		configPath := config.GetConfigPath()
		cfg = config.NewConfig(configPath, model.NewManifest())
		err := pathutil.EnsurePath(baseDir)
		if err != nil {
			log.Error(err)
			return -1
		}

		log.Highlight("nostromo config created")
	} else {
		log.Highlight("nostromo config exists, updating")
	}

	err := saveConfig(cfg, true)
	if err != nil {
		log.Error(err)
		return -1
	}

	return 0
}

// DestroyConfig deletes nostromo config file
func DestroyConfig() int {
	cfg := checkConfig()
	if cfg == nil {
		return -1
	}

	err := cfg.Delete()
	if err != nil {
		log.Error(err)
		return -1
	}

	log.Highlight("nostromo config deleted")

	err = shell.Commit(model.NewManifest())
	if err != nil {
		log.Error(err)
		return -1
	}

	return 0
}

// ShowConfig for nostromo config file
func ShowConfig(asJSON bool, asYAML bool, asTree bool) int {
	cfg := checkConfig()
	if cfg == nil {
		return -1
	}

	m := cfg.Manifest()

	if asJSON || asYAML {
		log.Bold("[manifest]")
		if asJSON {
			log.Regular(m.AsJSON())
		} else if asYAML {
			log.Regular(m.AsYAML())
		}
	} else if asTree {
		tree.PrintHr(m)
	} else {
		log.Bold("[manifest]")
		logFields(m, m.Config.Verbose)

		log.Bold("\n[config]")
		logFields(m.Config, m.Config.Verbose)

		if len(m.Commands) > 0 {
			log.Bold("\n[commands]")
			for _, cmd := range m.Commands {
				cmd.Walk(func(c *model.Command, s *bool) {
					logFields(c, m.Config.Verbose)
					if m.Config.Verbose {
						log.Regular()
					}
				})
			}
		} else if m.Config.Verbose {
			log.Regular()
		}

		if !m.Config.Verbose {
			log.Regular()
		}

		lines, err := shell.InitFileLines()
		if err != nil {
			return -1
		}

		log.Bold("[profile]")
		if len(lines) > 0 {
			log.Regular(strings.TrimSpace(lines))
		} else {
			log.Regular("empty")
		}
	}

	return 0
}

// SetConfig updates properties for nostromo settings
func SetConfig(key, value string) int {
	cfg := checkConfig()
	if cfg == nil {
		return -1
	}

	err := cfg.Set(key, value)
	if err != nil {
		log.Error(err)
		return -1
	}

	err = saveConfig(cfg, false)
	if err != nil {
		log.Error(err)
		return -1
	}

	return 0
}

// GetConfig reads properties from nostromo settings
func GetConfig(key string) int {
	cfg := checkConfig()
	if cfg == nil {
		return -1
	}

	log.Highlight(cfg.Get(key))
	return 0
}

// GenerateCompletions for all manifest commands and nostromo itself.
func GenerateCompletions(cmd *cobra.Command) int {
	// Generate completions for nostromo
	s, err := shell.Completion(cmd)
	if err != nil {
		return -1
	}
	log.Print(s)

	cfg := checkConfigQuiet()
	if cfg == nil {
		return 0
	}

	// Generate completions for manifest commands
	completions, err := shell.ManifestCompletion(cfg.Manifest())
	if err != nil {
		return 0
	}

	for _, completion := range completions {
		log.Print(completion)
	}

	return 0
}

// AddInteractive adds a command or substitution through user prompts
func AddInteractive() int {
	log.Highlight("Awesome! Let's add a new command or substitution to nostromo.")
	log.Regularf("Follow the prompts below to get started.\n\n")

	log.Regular("A nostromo command is a powerful shell alias that can be used to run one or many scoped commands.\n" +
		"Substitutions are scoped so sub commands all inherit them. You can have nostromo swap arguments on the command\n" +
		"line when it sees them simplifying your workflow.")
	isCmd := prompt.Choose("Choose what you would like to add (command)", []string{"command", "substitution"}, 0) == 0
	if !isCmd {
		log.Highlight("\nGreat, let's add a substitution to an existing command.\n")
	} else {
		log.Highlight("\nGreat, let's build a new command.\n")
	}

	if isCmd {
		log.Regular("A key path is a dot '.' delimited path to where you want to add your command.\n" +
			"Leave this blank if you want to add this to the root of the command tree.\n")
		keypath := prompt.String("Enter a key path (e.g., 'foo.bar') to attach your command (root)", "")

		log.Regular("\nBy default the command you provide is expected to run in the shell. However, nostromo provides the\n" +
			"functionality to run code as well in some popular languages.")
		languages := shell.SupportedLanguages()
		language := languages[prompt.Choose("Choose a language to run your command (sh)", languages, 0)]
		var cmd string
		var snippet string
		if language == "sh" {
			cmd = prompt.StringRequired("Enter the shell command (e.g., 'echo foo') to run")
		} else {
			snippet = prompt.StringRequired("Enter a single line code snippet to run")
		}
		alias := prompt.StringRequired("Enter the alias or shortcut (e.g., 'foo') to use")
		description := prompt.String("Enter a description (e.g., 'prints foo') for your command", "")

		log.Regular("\nCommands added to nostromo can be composed to build declarative tools effectively. However, if you\n" +
			"just want to use a boring old alias you can do that as well and manage them from nostromo.\n")
		aliasOnly := prompt.Confirm("Create a standard alias - say no :) (y/N)", false)
		var mode string
		if !aliasOnly {
			log.Highlight("\nGlad to see you're creating a nostromo command.\n")
			log.Regular("Now you can run this command in several different modes. By default, commands are \"concatenated\"\n" +
				"together so as nostromo walks the key path it builds up a final command to run.\n\n" +
				"However, you can choose to run a command \"independently\", which effectively adds a ';' after the command\n" +
				"to indicate to the shell to run separately. Or even \"exclusively\" which will ignore parent commands\n" +
				"and only run this one. The flexibility is provided to meet most needs.")
			modes := model.SupportedModes()
			mode = modes[prompt.Choose("Choose a command mode to use (concatenate)", modes, 0)]
		}
		if len(keypath) == 0 {
			keypath = alias
		} else {
			keypath = strings.Join([]string{keypath, alias}, ".")
		}
		log.Highlight("\nCreating command...\n")

		return AddCommand(keypath, cmd, description, snippet, language, aliasOnly, mode, false)
	}

	log.Regularf("A key path is a dot '.' delimited path to where you want to add your command.\n")
	log.Regular("Substitutions are added to an existing command node so you must specify where you would like to add it.\n")
	keypath := prompt.StringRequired("Enter a key path (e.g., 'foo.bar') to attach your command")

	log.Regular("\nThe original value is generally a longer string that you want to substitute with a shorthand version.\n" +
		"These might be long arguments that you want to sub out when you invoke your commands simplifying the call.\n")
	sub := prompt.StringRequired("Enter the original value")

	log.Regular("\nThe substitution is the value you will use on the CLI to run your nostromo command.\n" +
		"It's the shorter version of what you want nostromo to swap out into the actual call.\n")
	alias := prompt.StringRequired("Enter the substitution")
	log.Highlight("\nAdding substitution...\n")

	return AddSubstitution(keypath, sub, alias)
}

// AddCommand to the manifest
func AddCommand(keyPath, command, description, code, language string, aliasOnly bool, mode string, update bool) int {
	cfg := checkConfig()
	if cfg == nil {
		return -1
	}

	m := cfg.Manifest()

	if update && m.Find(keyPath) == nil {
		log.Error("no matching command found to update")
		return -1
	}

	snippet := &model.Code{
		Language: language,
		Snippet:  code,
	}

	_, err := m.AddCommand(keyPath, command, description, snippet, aliasOnly, mode)
	if err != nil {
		log.Error(err)
		return -1
	}

	cmd := m.Find(keyPath)
	if cmd == nil {
		log.Error("unable to find newly created command")
		return -1
	}

	err = saveConfig(cfg, false)
	if err != nil {
		log.Error(err)
		return -1
	}

	logFields(cmd, m.Config.Verbose)
	return 0
}

// RemoveCommand from the manifest
func RemoveCommand(keyPath string) int {
	cfg := checkConfig()
	if cfg == nil {
		return -1
	}

	_, err := cfg.Manifest().RemoveCommand(keyPath)
	if err != nil {
		log.Error(err)
		return -1
	}

	err = saveConfig(cfg, false)
	if err != nil {
		log.Error(err)
		return -1
	}

	return 0
}

// AddSubstitution to the manifest
func AddSubstitution(keyPath, name, alias string) int {
	cfg := checkConfig()
	if cfg == nil {
		return -1
	}

	m := cfg.Manifest()

	err := m.AddSubstitution(keyPath, name, alias)
	if err != nil {
		log.Error(err)
		return -1
	}

	err = saveConfig(cfg, false)
	if err != nil {
		log.Error(err)
		return -1
	}

	logFields(m.Find(keyPath), m.Config.Verbose)
	return 0
}

// RemoveSubstitution from the manifest
func RemoveSubstitution(keyPath, alias string) int {
	cfg := checkConfig()
	if cfg == nil {
		return -1
	}

	err := cfg.Manifest().RemoveSubstitution(keyPath, alias)
	if err != nil {
		log.Error(err)
		return -1
	}

	err = saveConfig(cfg, false)
	if err != nil {
		log.Error(err)
		return -1
	}

	return 0
}

// EvalString returns a command that can be used with `eval`
func EvalString(args []string) int {
	log.SetEcho(true)

	cfg := checkConfig()
	if cfg == nil {
		return -1
	}

	m := cfg.Manifest()

	language, cmd, err := m.ExecutionString(stringutil.SanitizeArgs(args))
	if err != nil {
		log.Error(err)
		return -1
	}

	cmdStr, err := shell.EvalString(cmd, language, m.Config.Verbose)
	if err != nil {
		log.Error(err)
		return -1
	}

	log.Print(cmdStr)
	return 0
}

// Find matching commands and substitutions
func Find(name string) int {
	cfg := checkConfig()
	if cfg == nil {
		return -1
	}

	m := cfg.Manifest()

	var matchingCmds []*model.Command
	var matchingSubs []*model.Command

	for _, cmd := range m.Commands {
		cmd.Walk(func(c *model.Command, s *bool) {
			if stringutil.ContainsCaseInsensitive(c.Name, name) || stringutil.ContainsCaseInsensitive(c.Alias, name) {
				matchingCmds = append(matchingCmds, c)
			}
			for _, sub := range c.Subs {
				if stringutil.ContainsCaseInsensitive(sub.Name, name) || stringutil.ContainsCaseInsensitive(sub.Alias, name) {
					matchingSubs = append(matchingSubs, c)
				}
			}
		})
	}

	if len(matchingCmds) == 0 && len(matchingSubs) == 0 {
		log.Highlight("no matching commands or substitutions found")
		return -1
	}

	log.Regular("[commands]")
	for _, cmd := range matchingCmds {
		logFields(cmd, m.Config.Verbose)
		if m.Config.Verbose {
			log.Regular()
		}
	}

	if !m.Config.Verbose {
		log.Regular()
	}
	log.Regular("[substitutions]")
	for _, cmd := range matchingSubs {
		logFields(cmd, m.Config.Verbose)
		if m.Config.Verbose {
			log.Regular()
		}
	}

	return 0
}

func checkConfigQuiet() *config.Config {
	return checkConfigCommon(true)
}

func checkConfig() *config.Config {
	return checkConfigCommon(false)
}

func checkConfigCommon(quiet bool) *config.Config {
	configPath := config.GetConfigPath()
	cfg, err := config.Parse(configPath)
	if err != nil {
		if !quiet {
			log.Error(err)
			log.Info("unable to open config file, be sure to run `nostromo init` if you haven't already")
		}
		return nil
	}

	log.SetVerbose(cfg.Manifest().Config.Verbose)

	return cfg
}

func saveConfig(cfg *config.Config, commit bool) error {
	m := cfg.Manifest()
	m.Version = ver.SemVer

	err := cfg.Save()
	if err != nil {
		return err
	}

	if commit {
		err = shell.Commit(m)
		if err != nil {
			return err
		}
	}

	return nil
}

func logFields(mapper log.FieldMapper, verbose bool) {
	if verbose {
		log.Table(mapper)
		return
	}
	log.Fields(mapper)
}
