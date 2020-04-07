package task

import (
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
	"strings"
)

var ver *version.Info

// SetVersion should be called before any task to ensure manifest is updated
func SetVersion(v *version.Info) {
	ver = v
}

// InitConfig of nostromo config file if not already initialized
func InitConfig() {
	cfg := checkConfigQuiet()

	if cfg == nil {
		cfg = config.NewConfig(config.Path, model.NewManifest())
		err := pathutil.EnsurePath("~/.nostromo")
		if err != nil {
			log.Error(err)
			return
		}

		log.Highlight("nostromo config created")
	} else {
		log.Highlight("nostromo config exists, updating")
	}

	err := saveConfig(cfg)
	if err != nil {
		log.Error(err)
	}
}

// DestroyConfig deletes nostromo config file
func DestroyConfig() {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	err := cfg.Delete()
	if err != nil {
		log.Error(err)
		return
	}

	log.Highlight("nostromo config deleted")

	err = shell.Commit(model.NewManifest())
	if err != nil {
		log.Error(err)
		return
	}
}

// ShowConfig for nostromo config file
func ShowConfig(asJSON bool, asYAML bool, asTree bool) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	m := cfg.Manifest()

	if asJSON || asYAML {
		log.Bold("[manifest]")
		if asJSON {
			log.Regular(m.AsJSON())
			log.Regular()
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
	}

	lines, err := shell.InitFileLines()
	if err != nil {
		return
	}

	log.Bold("[profile]")
	if len(lines) > 0 {
		log.Regular(strings.TrimSpace(lines))
	} else {
		log.Regular("empty")
	}
}

// SetConfig updates properties for nostromo settings
func SetConfig(key, value string) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	err := cfg.Set(key, value)
	if err != nil {
		log.Error(err)
		return
	}

	err = saveConfig(cfg)
	if err != nil {
		log.Error(err)
	}
}

// GetConfig reads properties from nostromo settings
func GetConfig(key string) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	log.Highlight(cfg.Get(key))
}

func GenerateCompletions(cmd *cobra.Command) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	// Generate completions for nostromo
	s, err := shell.Completion(cmd)
	if err != nil {
		log.Error(err)
	}
	log.Print(s)

	// Generate completions for manifest commands
	completions, err := shell.ManifestCompletion(cfg.Manifest())
	if err != nil {
		log.Error(err)
		return
	}

	for _, completion := range completions {
		log.Print(completion)
	}
}

// AddInteractive adds a command or substitution through user prompts
func AddInteractive() {
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

	var keypath string
	if isCmd {
		log.Regular("A key path is a dot '.' delimited path to where you want to add your command.\n" +
			"Leave this blank if you want to add this to the root of the command tree.\n")
		keypath = prompt.String("Enter a key path (e.g., 'foo.bar') to attach your command (root)", "")

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
			"just want to use a plain old alias you can do that as well and manage them from nostromo.\n")
		aliasOnly := prompt.Confirm("Create a standard alias only or make magic (y/N)", false)
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
		AddCommand(keypath, cmd, description, snippet, language, aliasOnly, mode)
	} else {
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
		AddSubstitution(keypath, sub, alias)
	}
}

// AddCommand to the manifest
func AddCommand(keyPath, command, description, code, language string, aliasOnly bool, mode string) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	m := cfg.Manifest()

	snippet := &model.Code{
		Language: language,
		Snippet:  code,
	}

	err := m.AddCommand(keyPath, command, description, snippet, aliasOnly, mode)
	if err != nil {
		log.Error(err)
		return
	}

	cmd := m.Find(keyPath)
	if cmd == nil {
		log.Error("unable to find newly created command")
		return
	}

	err = saveConfig(cfg)
	if err != nil {
		log.Error(err)
	}

	logFields(cmd, m.Config.Verbose)
}

// RemoveCommand from the manifest
func RemoveCommand(keyPath string) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	err := cfg.Manifest().RemoveCommand(keyPath)
	if err != nil {
		log.Error(err)
		return
	}

	err = saveConfig(cfg)
	if err != nil {
		log.Error(err)
	}
}

// AddSubstitution to the manifest
func AddSubstitution(keyPath, name, alias string) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	m := cfg.Manifest()

	err := m.AddSubstitution(keyPath, name, alias)
	if err != nil {
		log.Error(err)
		return
	}

	err = saveConfig(cfg)
	if err != nil {
		log.Error(err)
		return
	}

	logFields(m.Find(keyPath), m.Config.Verbose)
}

// RemoveSubstitution from the manifest
func RemoveSubstitution(keyPath, alias string) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	err := cfg.Manifest().RemoveSubstitution(keyPath, alias)
	if err != nil {
		log.Error(err)
		return
	}

	err = saveConfig(cfg)
	if err != nil {
		log.Error(err)
	}
}

// EvalString returns a command that can be used with `eval`
func EvalString(args []string) {
	log.SetEcho(true)

	cfg := checkConfig()
	if cfg == nil {
		return
	}

	m := cfg.Manifest()

	language, cmd, err := m.ExecutionString(stringutil.SanitizeArgs(args))
	if err != nil {
		log.Error(err)
		return
	}

	cmdStr, err := shell.EvalString(cmd, language, m.Config.Verbose)
	if err != nil {
		log.Error(err)
	}

	log.Print(cmdStr)
}

// Find matching commands and substitutions
func Find(name string) {
	cfg := checkConfig()
	if cfg == nil {
		return
	}

	m := cfg.Manifest()

	matchingCmds := []*model.Command{}
	matchingSubs := []*model.Command{}

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
		return
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
}

func checkConfigQuiet() *config.Config {
	return checkConfigCommon(true)
}

func checkConfig() *config.Config {
	return checkConfigCommon(false)
}

func checkConfigCommon(quiet bool) *config.Config {
	cfg, err := config.Parse(config.Path)
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

func saveConfig(cfg *config.Config) error {
	m := cfg.Manifest()
	m.Version = ver.SemVer

	err := cfg.Save()
	if err != nil {
		return err
	}

	err = shell.Commit(m)
	if err != nil {
		return err
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
