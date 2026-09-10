package task

import (
	"fmt"
	"strings"

	"github.com/pokanop/nostromo/config"
	"github.com/pokanop/nostromo/log"
	"github.com/pokanop/nostromo/model"
	"github.com/pokanop/nostromo/shell"
)

// The functions in this file hold the business logic shared by the CLI
// tasks and the web UI. They return errors instead of exit codes and never
// log so callers decide how to report results.

// CommandUpdate holds the editable fields of a command
type CommandUpdate struct {
	Alias       string
	Name        string
	Description string
	Mode        string
	AliasOnly   bool
	Disabled    bool
	Language    string
	Snippet     string
}

// LoadConfig loads the nostromo config and applies its theme and verbosity
func LoadConfig() (*config.Config, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	log.SetTheme(cfg.Spaceport().Config.Theme)
	log.SetVerbose(cfg.Spaceport().Config.IsVerbose())

	return cfg, nil
}

// SaveConfig persists the spaceport and core manifest with a backup
func SaveConfig(cfg *config.Config) error {
	return saveConfig(cfg, false)
}

// AddCommandToConfig adds or updates a command in the core manifest and saves
func AddCommandToConfig(cfg *config.Config, keyPath, command, description, code, language string, aliasOnly bool, mode string, platforms []string, env model.EnvChanges, update bool) (*model.Command, error) {
	m := cfg.Spaceport().CoreManifest()

	if (cfg.Spaceport().Config.AliasesOnly || aliasOnly) && (len(env.Set) > 0 || len(env.Dotenv) > 0) {
		return nil, fmt.Errorf("env and dotenv are not supported for alias only commands")
	}

	if update {
		cmd := m.Find(keyPath)
		if cmd == nil {
			return nil, fmt.Errorf("no matching command found to update")
		}
		if len(command) == 0 {
			// Keep same command if not supplied
			command = cmd.Name
		}
		if platforms == nil {
			platforms = cmd.Platforms
		}
	}

	if err := validateCode(code, language); err != nil {
		return nil, err
	}
	if len(mode) > 0 && !model.IsModeSupported(mode) {
		return nil, fmt.Errorf("invalid mode, supported modes: %s", model.SupportedModes())
	}

	snippet := &model.Code{
		Language: language,
		Snippet:  code,
	}

	aliasOnly = cfg.Spaceport().Config.AliasesOnly || aliasOnly
	if len(mode) == 0 {
		mode = cfg.Spaceport().Config.Mode.String()
	}

	if _, err := m.AddCommand(keyPath, command, description, snippet, aliasOnly, mode, platforms); err != nil {
		return nil, err
	}

	cmd := m.Find(keyPath)
	if cmd == nil {
		return nil, fmt.Errorf("unable to find newly created command")
	}
	cmd.ApplyEnv(env)

	if err := saveConfig(cfg, false); err != nil {
		return nil, err
	}

	return cmd, nil
}

// UpdateCommandInConfig replaces the editable fields of an existing core
// manifest command, renaming it if the alias changed, and saves
func UpdateCommandInConfig(cfg *config.Config, keyPath string, upd CommandUpdate) (*model.Command, error) {
	m := cfg.Spaceport().CoreManifest()

	cmd := m.Find(keyPath)
	if cmd == nil {
		return nil, fmt.Errorf("command not found")
	}

	if err := validateCode(upd.Snippet, upd.Language); err != nil {
		return nil, err
	}
	if len(upd.Mode) > 0 && !model.IsModeSupported(upd.Mode) {
		return nil, fmt.Errorf("invalid mode, supported modes: %s", model.SupportedModes())
	}

	if len(upd.Alias) > 0 && upd.Alias != cmd.Alias {
		if err := m.RenameCommand(keyPath, upd.Alias, ""); err != nil {
			return nil, err
		}
	}

	cmd.Name = upd.Name
	cmd.Description = upd.Description
	cmd.AliasOnly = upd.AliasOnly
	cmd.Disabled = upd.Disabled
	cmd.Code = &model.Code{Language: upd.Language, Snippet: upd.Snippet}
	if len(upd.Mode) > 0 {
		cmd.Mode = model.ModeFromString(upd.Mode)
	}

	if err := saveConfig(cfg, false); err != nil {
		return nil, err
	}

	return cmd, nil
}

// RemoveCommandFromConfig removes a command from the core manifest and saves
func RemoveCommandFromConfig(cfg *config.Config, keyPath string) error {
	if _, err := cfg.Spaceport().CoreManifest().RemoveCommand(keyPath); err != nil {
		return err
	}

	return saveConfig(cfg, false)
}

// RenameCommandInConfig renames a command in the core manifest and saves
func RenameCommandInConfig(cfg *config.Config, keyPath, name, description string) error {
	if err := cfg.Spaceport().CoreManifest().RenameCommand(keyPath, name, description); err != nil {
		return err
	}

	return saveConfig(cfg, false)
}

// AddSubstitutionToConfig adds a substitution to a core manifest command and saves
func AddSubstitutionToConfig(cfg *config.Config, keyPath, name, alias string) (*model.Command, error) {
	if len(strings.TrimSpace(name)) == 0 || len(strings.TrimSpace(alias)) == 0 {
		return nil, fmt.Errorf("substitution requires both an original value and an alias")
	}

	m := cfg.Spaceport().CoreManifest()
	if err := m.AddSubstitution(keyPath, name, alias); err != nil {
		return nil, err
	}

	if err := saveConfig(cfg, false); err != nil {
		return nil, err
	}

	return m.Find(keyPath), nil
}

// RemoveSubstitutionFromConfig removes a substitution from a core manifest command and saves
func RemoveSubstitutionFromConfig(cfg *config.Config, keyPath, alias string) error {
	if err := cfg.Spaceport().CoreManifest().RemoveSubstitution(keyPath, alias); err != nil {
		return err
	}

	return saveConfig(cfg, false)
}

// FindMatches returns commands whose name, alias or key path match, and the
// commands owning substitutions that match, across all manifests
func FindMatches(cfg *config.Config, name string) ([]*model.Command, []*model.Command) {
	cmds, subs := cfg.Spaceport().Search(name)
	matchingCmds := make([]*model.Command, 0, len(cmds))
	for _, r := range cmds {
		matchingCmds = append(matchingCmds, r.Command)
	}
	matchingSubs := []*model.Command{}
	seen := map[*model.Command]bool{}
	for _, r := range subs {
		if !seen[r.Command] {
			seen[r.Command] = true
			matchingSubs = append(matchingSubs, r.Command)
		}
	}
	return matchingCmds, matchingSubs
}

func validateCode(code, language string) error {
	if len(code) == 0 && len(language) == 0 {
		return nil
	}
	if len(code) == 0 || len(language) == 0 {
		return fmt.Errorf("code snippet and language must both be provided")
	}
	if !shell.IsSupportedLanguage(language) {
		return fmt.Errorf("invalid language %s, must be in [%s]", language, strings.Join(shell.SupportedLanguages(), ","))
	}
	return nil
}
