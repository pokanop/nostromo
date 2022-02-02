package task

import (
	"os"
	"path/filepath"
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

// InitConfig of nostromo config file if not already initialized
func InitConfig(cmd *cobra.Command) int {
	// Attempt to load existing config
	cfg, err := config.LoadConfig()
	if err != nil {
		// Create a new config
		cfg, err = config.NewConfig()
		if err != nil {
			log.Error(err)
			return -1
		}
		log.Highlight("nostromo config created")
	} else {
		log.Highlight("nostromo config exists, updating")
	}

	err = saveConfig(cfg, true)
	if err != nil {
		log.Error(err)
		return -1
	}

	// Generate completion files
	GenerateCompletions("bash", cmd, true)
	GenerateCompletions("zsh", cmd, true)
	GenerateCompletions("fish", cmd, true)
	GenerateCompletions("powershell", cmd, true)

	return 0
}

// DestroyConfig for core manifest file
func DestroyConfig(nuke bool) int {
	if nuke {
		err := os.RemoveAll(pathutil.Abs(config.BaseDir()))
		if err != nil {
			return -1
		}
		return 0
	}

	cfg := checkConfig()
	if cfg == nil {
		return -1
	}

	err := cfg.DeleteManifest(model.CoreManifestName)
	if err != nil {
		log.Error(err)
		return -1
	}

	log.Highlight("nostromo config destroyed")

	m, err := config.NewCoreManifest()
	if err != nil {
		log.Error(err)
		return -1
	}

	err = shell.Commit(m)
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

	verbose := cfg.Spaceport().CoreManifest().Config.Verbose
	for i, m := range cfg.Spaceport().Manifests() {
		if i > 0 {
			log.Regular()
		}
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
			logFields(m, verbose)

			if m.IsCore() {
				log.Bold("\n[config]")
				logFields(m.Config, verbose)
			}

			if len(m.Commands) > 0 {
				log.Bold("\n[commands]")
				for _, cmd := range m.Commands {
					cmd.Walk(func(c *model.Command, s *bool) {
						logFields(c, verbose)
						if verbose {
							log.Regular()
						}
					})
				}
			} else if verbose {
				log.Regular()
			}

			if !verbose {
				log.Regular()
			}

			lines := shell.InitFileLines()
			if m.IsCore() {
				log.Bold("[profile]")
				if len(lines) > 0 {
					log.Regular(strings.TrimSpace(lines))
				} else {
					log.Regular("empty")
				}
			}
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

// FetchCommands builds the nostromo command tree and returns the top level
// commands for execution
func FetchCommands() []*cobra.Command {
	cmds := []*cobra.Command{}

	cfg := checkConfigQuiet()
	if cfg == nil {
		return cmds
	}

	for _, cmd := range cfg.Spaceport().Commands() {
		cmds = append(cmds, cmd.CobraCommand())
	}

	return cmds
}

// GenerateCompletions for all manifest commands and nostromo itself.
func GenerateCompletions(sh string, cmd *cobra.Command, writeFile bool) int {
	// Generate completions for nostromo
	s, err := shell.Completion(sh, cmd)
	if err != nil {
		return -1
	}
	if !writeFile {
		log.Print(s)
	}

	cfg := checkConfigQuiet()
	if cfg == nil {
		return 0
	}

	// Generate completions for manifest commands
	completions, err := shell.SpaceportCompletion(sh, cfg.Spaceport())
	if err != nil {
		return 0
	}

	for _, completion := range completions {
		if !writeFile {
			log.Print(completion)
		}
		s += "\n" + completion
	}

	if writeFile {
		err = config.WriteCompletion(sh, s)
		if err != nil {
			log.Warningf("unable to write completion file for %s\n", sh)
		}
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

	m := cfg.Spaceport().CoreManifest()

	if update && m.Find(keyPath) == nil {
		log.Error("no matching command found to update")
		return -1
	}

	snippet := &model.Code{
		Language: language,
		Snippet:  code,
	}

	aliasOnly = m.Config.AliasesOnly || aliasOnly
	if len(mode) == 0 {
		mode = m.Config.Mode.String()
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

	_, err := cfg.Spaceport().CoreManifest().RemoveCommand(keyPath)
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

	m := cfg.Spaceport().CoreManifest()

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

	err := cfg.Spaceport().CoreManifest().RemoveSubstitution(keyPath, alias)
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

	var cmdStr string
	var err error
	for _, m := range cfg.Spaceport().Manifests() {
		var language, cmd string
		language, cmd, err = m.ExecutionString(args)
		if err != nil {
			continue
		}

		cmdStr, err = shell.EvalString(cmd, language, m.Config.Verbose)
		if err != nil {
			continue
		}
		break
	}

	if len(cmdStr) == 0 && err != nil {
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

	var matchingCmds []*model.Command
	var matchingSubs []*model.Command

	for _, m := range cfg.Spaceport().Manifests() {
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
	}

	if len(matchingCmds) == 0 && len(matchingSubs) == 0 {
		log.Highlight("no matching commands or substitutions found")
		return -1
	}

	m := cfg.Spaceport().CoreManifest()

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

func Sync(force bool, sources []string) int {
	cfg := checkConfig()
	if cfg == nil {
		return -1
	}

	if err := cfg.Sync(force, sources); err != nil {
		log.Error(err)
		return -1
	}

	if len(sources) == 0 {
		log.Highlight("synchronized nostromo manifests")
	} else {
		log.Highlight("docked nostromo manifests")
	}

	return 0
}

func Detach(name string, keyPaths []string, targetKeyPath, description string, keepOriginal bool) int {
	cfg := checkConfig()
	if cfg == nil {
		return -1
	}
	s := cfg.Spaceport()

	// First try to find the relevant keypaths
	cmds := []*model.Command{}
	for _, keyPath := range keyPaths {
		cmd := s.FindCommand(keyPath)
		if cmd == nil {
			log.Errorf("keypath not found: %s\n", keyPath)
			return -1
		}
		cmds = append(cmds, cmd)
	}

	// Name is the output file name, check if manifest already exists
	var m *model.Manifest
	var err error
	path := filepath.Join(pathutil.Abs(config.BaseDir()), config.DefaultManifestsDir, name+".yaml")
	if m, err = config.Parse(path); err != nil {
		// Manifest does not exist, create a new one
		m = config.NewManifest(name)
	}

	// Merge or add commands to manifest
	m.ImportCommands(cmds, targetKeyPath, description)

	// Remove original node if needed, only applies to core manifest
	if !keepOriginal {
		cm := s.CoreManifest()
		for _, keyPath := range keyPaths {
			_, err = cm.RemoveCommand(keyPath)
			if err != nil {
				log.Warningf("cannot remove %s: %s\n", keyPath, err)
			}
		}

		// Save core manifest updates
		err = config.SaveManifest(cm, true)
		if err != nil {
			log.Error(err)
			return -1
		}
	}

	// Update spaceport and save manifest
	s.AddManifest(m)
	err = config.SaveManifest(m, false)
	if err != nil {
		log.Error(err)
		return -1
	}
	err = config.SaveSpaceport(s)
	if err != nil {
		log.Error(err)
		return -1
	}

	return 0
}

func RegenerateID(name string) int {
	cfg := checkConfig()
	var m *model.Manifest
	if len(name) > 0 {
		m = cfg.Spaceport().FindManifest(name)
		if m == nil {
			log.Errorf("no manifest named %s exists\n", name)
			return -1
		}
	} else {
		m = cfg.Spaceport().CoreManifest()
	}

	v := version.NewInfo(m.Version.SemVer, m.Version.GitCommit, m.Version.BuildDate)
	m.Version.Update(v)

	err := config.SaveManifest(m, m.IsCore())
	if err != nil {
		log.Error(err)
		return -1
	}

	return 0
}

// Undock a manifest from nostromo installation
func Undock(name string) int {
	if len(name) == 0 {
		log.Errorf("no manifest named %s found\n", name)
		return -1
	}

	cfg := checkConfig()
	if cfg == nil {
		return -1
	}

	err := cfg.DeleteManifest(name)
	if err != nil {
		log.Error(err)
		return -1
	}

	if !cfg.Spaceport().RemoveManifest(name) {
		log.Warningf("spaceport missing manifest %s\n", name)
	}

	err = config.SaveSpaceport(cfg.Spaceport())
	if err != nil {
		log.Error(err)
		return -1
	}

	log.Highlight("undocked nostromo manifest")

	return 0
}

func checkConfigQuiet() *config.Config {
	return checkConfigCommon(true)
}

func checkConfig() *config.Config {
	return checkConfigCommon(false)
}

func checkConfigCommon(quiet bool) *config.Config {
	cfg, err := config.LoadConfig()
	if err != nil {
		if !quiet {
			log.Error(err)
			log.Info("unable to open config file, be sure to run `nostromo init` if you haven't already")
		}
		return nil
	}

	log.SetVerbose(cfg.Spaceport().CoreManifest().Config.Verbose)

	return cfg
}

func saveConfig(cfg *config.Config, commit bool) error {
	m := cfg.Spaceport().CoreManifest()

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
